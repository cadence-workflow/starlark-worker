package activity

import (
	"context"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.uber.org/zap"
)

func GetLogger(ctx context.Context) *zap.Logger {
	if b, ok := workflow.GetBackend(ctx); ok {
		return b.GetActivityLogger(ctx)
	}
	return nil
}
