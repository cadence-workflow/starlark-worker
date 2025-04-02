package activity

import (
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.uber.org/zap"
)

type Activity interface {
	GetLogger(ctx workflow.Context) *zap.Logger
}

var backend Activity

func RegisterBackend(b Activity) {
	backend = b
}

func GetLogger(ctx workflow.Context) *zap.Logger {
	if backend == nil {
		panic("workflow backend not registered")
	}
	return backend.GetLogger(ctx)
}
