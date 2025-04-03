package activity

import (
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.uber.org/zap"
)

type Activity interface {
	GetLogger(ctx workflow.Context) *zap.Logger
}

func GetLogger(ctx workflow.Context) *zap.Logger {
	if b, ok := workflow.GetBackend(ctx); ok {
		return b.GetLogger(ctx)
	}
	return nil
}
