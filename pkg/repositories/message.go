package repositories

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/meesooqa/tgtag/internal/db"
	"github.com/meesooqa/tgtag/pkg/models"
)

type MessageRepository struct {
	log        *slog.Logger
	collection *mongo.Collection
}

func NewMessageRepository(log *slog.Logger, db *db.MongoDB) *MessageRepository {
	return &MessageRepository{
		log:        log,
		collection: db.GetCollectionMessages(),
	}
}

func (r *MessageRepository) UpsertMany(messagesChan <-chan models.Message) {
	batchSize := 10
	flushPeriod := 2 // Seconds

	s := newSaver(r.log, r.collection, batchSize, time.Duration(flushPeriod)*time.Second, 50)
	go func() {
		for msg := range messagesChan {
			doc := bson.M{
				"message_id": msg.MessageID,
				"datetime":   msg.Datetime,
				"group":      msg.Group,
				"uuid":       msg.UUID,
				"tags":       msg.Tags,
			}
			if err := s.Save(doc); err != nil {
				r.log.Error("Saver error", "err", err)
			}
		}
	}()
	time.Sleep(time.Duration(flushPeriod+1) * time.Second) // wait flushPeriod
	s.Close()
	r.log.Debug("all data has been successfully saved to MongoDB")
}

func (r *MessageRepository) Find(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*models.Message, error) {
	cursor, err := r.collection.Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var items []*models.Message
	for cursor.Next(ctx) {
		var item models.Message
		if err := cursor.Decode(&item); err != nil {
			return nil, err
		}
		items = append(items, &item)
	}
	if err := cursor.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (r *MessageRepository) GetUniqueValues(ctx context.Context, fieldName string) ([]string, error) {
	values, err := r.collection.Distinct(ctx, fieldName, bson.D{})
	if err != nil {
		return nil, fmt.Errorf("distinct failed: %w", err)
	}
	return convertToStrings(values), nil
}

func (r *MessageRepository) GetTags(ctx context.Context, query string) ([]string, error) {
	var results []string
	cursor, err := r.collection.Aggregate(ctx,
		mongo.Pipeline{
			bson.D{{"$unwind", "$tags"}},
			bson.D{{"$group", bson.D{{"_id", "$tags"}}}},
			bson.D{{"$match", bson.D{{"_id", bson.D{{"$regex", query}, {"$options", "i"}}}}}},
		})
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := cursor.Close(ctx); err != nil {
			r.log.Error("cursor close error", "err", err)
		}
	}()

	type res struct {
		ID string `bson:"_id"`
	}
	for cursor.Next(ctx) {
		var elem res
		if err := cursor.Decode(&elem); err != nil {
			return nil, fmt.Errorf("decode error: %w", err)
		}
		if elem.ID != "" {
			results = append(results, elem.ID)
		}
	}

	return results, nil
}

func convertToStrings(values []interface{}) []string {
	result := make([]string, 0, len(values))
	for _, v := range values {
		if str, ok := v.(string); ok {
			result = append(result, str)
		}
	}
	return result
}
