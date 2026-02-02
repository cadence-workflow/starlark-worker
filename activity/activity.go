package activity

import (
	"context"
	"github.com/cadence-workflow/starlark-worker/internal"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.uber.org/zap"
)

type (
	Info = internal.ActivityInfo
)

func GetLogger(ctx context.Context) *zap.Logger {
	if b, ok := workflow.GetBackend(ctx); ok {
		return b.GetActivityLogger(ctx)
	}
	return nil
}

func GetInfo(ctx context.Context) Info {
	if b, ok := workflow.GetBackend(ctx); ok {
		return b.GetActivityInfo(ctx)
	}
	return Info{}
}
