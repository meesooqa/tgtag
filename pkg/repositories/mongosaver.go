package repositories

import (
	"context"
	"errors"
	"log/slog"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Inserter представляет сущность, поддерживающую пакетную вставку документов.
type inserter interface {
	BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error)
}

// Saver отвечает за сбор и пакетную отправку данных в MongoDB.
type saver struct {
	log         *slog.Logger
	collection  inserter
	dataChan    chan bson.M
	batchSize   int
	flushPeriod time.Duration
	wg          sync.WaitGroup
	mu          sync.Mutex
	closed      bool
}

// NewSaver создаёт новый Saver с указанными параметрами.
func newSaver(log *slog.Logger, collection inserter, batchSize int, flushPeriod time.Duration, bufferSize int) *saver {
	s := &saver{
		log:         log,
		collection:  collection,
		dataChan:    make(chan bson.M, bufferSize),
		batchSize:   batchSize,
		flushPeriod: flushPeriod,
	}
	s.wg.Add(1)
	go s.run()
	return s
}

// run запускает обработку канала и периодическое сохранение в MongoDB.
func (s *saver) run() {
	defer s.wg.Done()
	ticker := time.NewTicker(s.flushPeriod)
	defer ticker.Stop()

	var batch []bson.M

	for {
		select {
		case doc, ok := <-s.dataChan:
			if !ok {
				if len(batch) > 0 {
					s.saveBatch(batch)
				}
				return
			}
			batch = append(batch, doc)

			if len(batch) >= s.batchSize {
				s.saveBatch(batch)
				batch = batch[:0]
			}
		case <-ticker.C:
			if len(batch) > 0 {
				s.saveBatch(batch)
				batch = batch[:0]
			}
		}
	}
}

// saveBatch сохраняет батч документов в MongoDB.
// 1) Если документа с UUID нет – вставляем новый.
// 2) Если документ с UUID уже есть:
//   - Если tags отличаются – обновляем поле tags (и, например, datetime).
//   - Если tags совпадают – обновление производится, но фактически документ не меняется.
func (s *saver) saveBatch(batch []bson.M) {
	var models []mongo.WriteModel

	for _, doc := range batch {
		// Фильтр всегда ищет документ по UUID
		filter := bson.M{"uuid": doc["uuid"]}

		// Операция обновления:
		// - $set устанавливает поля (при обновлении, если tags изменились)
		// - $setOnInsert гарантирует, что при вставке будет заполнен UUID
		update := bson.M{
			"$set": bson.M{
				"message_id": doc["message_id"],
				"group":      doc["group"],
				"datetime":   doc["datetime"],
				"tags":       doc["tags"],
			},
			"$setOnInsert": bson.M{
				"uuid": doc["uuid"],
			},
		}

		// Используем UpdateOne с upsert:true.
		model := mongo.NewUpdateOneModel().
			SetFilter(filter).
			SetUpdate(update).
			SetUpsert(true)

		models = append(models, model)
	}

	opts := options.BulkWrite().SetOrdered(false)
	result, err := s.collection.BulkWrite(context.TODO(), models, opts)
	s.log.Debug("BulkWrite result", "result", result)
	if err != nil {
		s.log.Error("BulkWrite failed", "err", err)
	}
}

// Save добавляет документ в очередь сохранения.
func (s *saver) Save(doc bson.M) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return errors.New("saver is closed")
	}
	s.dataChan <- doc
	return nil
}

// Close завершает работу и сохраняет остатки.
func (s *saver) Close() {
	s.mu.Lock()
	if !s.closed {
		s.closed = true
		close(s.dataChan)
	}
	s.mu.Unlock()
	s.wg.Wait()
}
