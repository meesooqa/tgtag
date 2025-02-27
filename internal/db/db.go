package db

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/meesooqa/tgtag/internal/config"
)

type MongoDB struct {
	log    *slog.Logger
	Conf   *config.MongoConfig
	client *mongo.Client
}

func NewMongoDB(log *slog.Logger, conf *config.MongoConfig) *MongoDB {
	return &MongoDB{
		log:  log,
		Conf: conf,
	}
}

func (db *MongoDB) Init() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	connectedClient, err := mongo.Connect(ctx, options.Client().ApplyURI(db.Conf.URI))
	if err != nil {
		return err
	}
	err = connectedClient.Ping(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to ping MongoDB: %v", err)
	}

	db.client = connectedClient

	if err := db.createUniqueUuidIndex(context.TODO()); err != nil {
		db.log.Error("creating index", "err", err)
	}

	return nil
}

func (db *MongoDB) Close() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if db.client != nil {
		if err := db.client.Disconnect(ctx); err != nil {
			db.log.Error("failed to disconnect MongoDB", "err", err)
		}
	}
}

func (db *MongoDB) GetDatabase() *mongo.Database {
	return db.client.Database(db.Conf.Database)
}

func (db *MongoDB) GetCollection(collectionName string) *mongo.Collection {
	return db.GetDatabase().Collection(collectionName)
}

func (db *MongoDB) GetCollectionMessages() *mongo.Collection {
	return db.GetDatabase().Collection(db.Conf.CollectionMessages)
}

func (db *MongoDB) createUniqueUuidIndex(ctx context.Context) error {
	collection := db.GetCollectionMessages()
	indexModel := mongo.IndexModel{
		Keys:    bson.D{{Key: "uuid", Value: 1}},
		Options: options.Index().SetUnique(true),
	}
	_, err := collection.Indexes().CreateOne(ctx, indexModel)
	return err
}
