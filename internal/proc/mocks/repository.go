package mocks

import (
	"context"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/meesooqa/tgtag/pkg/models"
)

type RepositoryMock struct {
	UpsertCalls []models.Message
	Err         error
}

func (f *RepositoryMock) UpsertMany(messagesChan <-chan models.Message) {
	for m := range messagesChan {
		f.UpsertCalls = append(f.UpsertCalls, m)
	}
}

func (f *RepositoryMock) GetGroups(ctx context.Context) ([]string, error) {
	return nil, nil
}

func (f *RepositoryMock) Find(ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]*models.Message, error) {
	return nil, nil
}
