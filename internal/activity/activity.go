package activity

import (
	"context"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.uber.org/zap"
)

type Activity interface {
	GetLogger(ctx context.Context) *zap.Logger
}

func GetLogger(ctx context.Context) *zap.Logger {
	if b, ok := workflow.GetBackend(ctx); ok {
		return b.GetActivityLogger(ctx)
	}
	return nil
}
