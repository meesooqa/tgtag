package analysis

import (
	"context"
	"log/slog"
)

type AnalyzedData any

type AnalyzedDataProvider interface {
	SetLogger(log *slog.Logger)
	GetData(ctx context.Context, group string) AnalyzedData
}
