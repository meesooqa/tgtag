package main_ext

import (
	"context"
	"log/slog"

	"github.com/meesooqa/tgtag/pkg/data"
	"github.com/meesooqa/tgtag/pkg/repositories"
)

type GroupDataProvider struct {
	log  *slog.Logger
	repo repositories.Repository
}

func NewGroupDataProvider(repo repositories.Repository) *GroupDataProvider {
	return &GroupDataProvider{
		repo: repo,
	}
}

func (p *GroupDataProvider) SetLogger(log *slog.Logger) {
	p.log = log
}

func (p *GroupDataProvider) GetData(ctx context.Context, group string) (data.Data, error) {
	result, err := p.repo.GetGroups(ctx)
	if err != nil {
		return nil, err
	}
	return result, nil
}
