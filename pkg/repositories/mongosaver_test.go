package repositories

import (
	"bytes"
	"context"
	"log/slog"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// BulkWriteCall хранит параметры вызова BulkWrite.
type bulkWriteCall struct {
	Ctx     context.Context
	Models  []mongo.WriteModel
	Options *options.BulkWriteOptions
}

// FakeInserter реализует интерфейс Inserter для тестирования.
type fakeInserter struct {
	mu    sync.Mutex
	Calls []bulkWriteCall
	// При необходимости можно симулировать ошибку.
	Err error
}

func (f *fakeInserter) BulkWrite(ctx context.Context, models []mongo.WriteModel, opts ...*options.BulkWriteOptions) (*mongo.BulkWriteResult, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var bulkOpts *options.BulkWriteOptions
	if len(opts) > 0 {
		bulkOpts = opts[0]
	}
	call := bulkWriteCall{
		Ctx:     ctx,
		Models:  models,
		Options: bulkOpts,
	}
	f.Calls = append(f.Calls, call)
	// Возвращаем фиктивный результат: считаем, что все операции upsert прошли успешно.
	result := &mongo.BulkWriteResult{
		MatchedCount:  0,
		ModifiedCount: 0,
		UpsertedCount: int64(len(models)),
	}
	return result, f.Err
}

func (f *fakeInserter) GetCalls() []bulkWriteCall {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.Calls
}

// TestSaver_BatchProcessing проверяет, что Saver корректно группирует документы,
// когда размер батча достигнут, и вызывает BulkWrite с ожидаемыми параметрами.
func TestSaver_BatchProcessing(t *testing.T) {
	fakeInserter := &fakeInserter{}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	// Устанавливаем batchSize = 2 и очень длинный flushPeriod, чтобы не срабатывать по таймеру.
	svr := newSaver(logger, fakeInserter, 2, 5*time.Second, 10)

	now := time.Now()
	doc1 := bson.M{
		"uuid":     "msg1",
		"tags":     []string{"tag1", "tag2"},
		"datetime": now,
	}
	doc2 := bson.M{
		"uuid":     "msg2",
		"tags":     []string{"tag3"},
		"datetime": now.Add(time.Minute),
	}

	if err := svr.Save(doc1); err != nil {
		t.Fatalf("Save failed: %v", err)
	}
	if err := svr.Save(doc2); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Закрываем Saver, чтобы сбросить оставшиеся документы.
	svr.Close()

	calls := fakeInserter.GetCalls()
	// Expected 1 BulkWrite call
	assert.Equal(t, 1, len(calls))

	// Ожидаем один вызов BulkWrite с 2-мя моделями.
	call := calls[0]
	// Expected 2 models in BulkWrite call
	assert.Equal(t, 2, len(call.Models))

	// Проверяем каждую модель.
	for i, model := range call.Models {
		updateModel, ok := model.(*mongo.UpdateOneModel)
		if !ok {
			t.Errorf("Model %d is not of type *mongo.UpdateOneModel", i)
			continue
		}
		// Фильтр должен искать по "UUID".
		filter, ok := updateModel.Filter.(bson.M)
		if !ok {
			t.Errorf("Model %d filter is not of type bson.M", i)
			continue
		}
		var expectedMsgID string
		var expectedTags []string
		var expectedDatetime time.Time
		if i == 0 {
			expectedMsgID = "msg1"
			expectedTags = []string{"tag1", "tag2"}
			expectedDatetime = now
		} else if i == 1 {
			expectedMsgID = "msg2"
			expectedTags = []string{"tag3"}
			expectedDatetime = now.Add(time.Minute)
		}
		assert.Equal(t, expectedMsgID, filter["uuid"], "expected filter UUID expectedMsgID")

		// Проверяем документ обновления.
		updateDoc, ok := updateModel.Update.(bson.M)
		if !ok {
			t.Errorf("Model %d update is not of type bson.M", i)
			continue
		}
		// Проверяем секцию $set.
		setPart, ok := updateDoc["$set"].(bson.M)
		if !ok {
			t.Errorf("Model %d: $set part missing or not a bson.M", i)
			continue
		}
		// Проверяем поле tags.
		tags, ok := setPart["tags"].([]string)
		if !ok {
			// Если тип не []string, возможно это bson.A, попробуем преобразовать.
			arr, ok2 := setPart["tags"].(bson.A)
			if !ok2 {
				t.Errorf("Model %d: tags is not []string or bson.A", i)
				continue
			}
			var strTags []string
			for _, v := range arr {
				if s, ok3 := v.(string); ok3 {
					strTags = append(strTags, s)
				}
			}
			tags = strTags
		}
		assert.ElementsMatch(t, expectedTags, tags)

		// Проверяем поле datetime.
		dt, ok := setPart["datetime"].(time.Time)
		if !ok {
			t.Errorf("Model %d: datetime is not time.Time", i)
		}
		assert.Equal(t, expectedDatetime, dt)

		// Проверяем секцию $setOnInsert.
		setOnInsert, ok := updateDoc["$setOnInsert"].(bson.M)
		if !ok {
			t.Errorf("Model %d: $setOnInsert missing or not a bson.M", i)
			continue
		}
		assert.Equal(t, expectedMsgID, setOnInsert["uuid"])

		// Проверяем, что Upsert установлен в true.
		assert.NotNil(t, updateModel.Upsert)
		assert.True(t, *updateModel.Upsert)
	}
}

// TestSaver_FlushPeriod проверяет, что если размер батча не достигнут, то срабатывает flushPeriod.
func TestSaver_FlushPeriod(t *testing.T) {
	fakeInserter := &fakeInserter{}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	// Устанавливаем batchSize = 10, flushPeriod короткий (например, 50мс) и bufferSize = 10.
	svr := newSaver(logger, fakeInserter, 10, 50*time.Millisecond, 10)

	now := time.Now()
	doc := bson.M{
		"uuid":     "msg_flush",
		"tags":     []string{"flushTag"},
		"datetime": now,
	}
	if err := svr.Save(doc); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Ждем, чтобы flushPeriod сработал.
	time.Sleep(100 * time.Millisecond)
	svr.Close()

	calls := fakeInserter.GetCalls()
	assert.GreaterOrEqual(t, 1, len(calls), "Expected at least 1 BulkWrite call due to flushPeriod, got 0")

	// Проверяем, что документ с "msg_flush" присутствует в одном из вызовов.
	found := false
	for _, call := range calls {
		for _, model := range call.Models {
			updateModel, ok := model.(*mongo.UpdateOneModel)
			if !ok {
				continue
			}
			filter, ok := updateModel.Filter.(bson.M)
			if !ok {
				continue
			}
			if filter["uuid"] == "msg_flush" {
				found = true
				break
			}
		}
	}
	assert.True(t, found, "Document with UUID 'msg_flush' not found in BulkWrite calls")
}

// TestSaver_SaveAfterClose проверяет, что вызов Save после закрытия Saver возвращает ошибку.
func TestSaver_SaveAfterClose(t *testing.T) {
	fakeInserter := &fakeInserter{}
	var buf bytes.Buffer
	logger := slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo}))
	svr := newSaver(logger, fakeInserter, 2, 5*time.Second, 10)
	svr.Close()

	err := svr.Save(bson.M{
		"uuid":     "msg_after_close",
		"tags":     []string{"tag"},
		"datetime": time.Now(),
	})
	if err == nil {
		t.Errorf("Expected error when saving after Close, got nil")
	}
}
