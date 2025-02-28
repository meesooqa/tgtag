package data

import (
	"context"
	"log/slog"
)

type Data any

type Provider interface {
	SetLogger(log *slog.Logger)
	GetData(ctx context.Context, group string) (Data, error)
}
