package repositories

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/meesooqa/tgtag/pkg/models"
)

type Repository interface {
	Find(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*models.Message, error)
	UpsertMany(messagesChan <-chan models.Message)
	GetGroups(ctx context.Context) ([]string, error)
}
