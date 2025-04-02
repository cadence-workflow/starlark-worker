package backend

import (
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"go.uber.org/zap"
)

type Backend interface {
	RegisterWorker(url string, domain string, taskList string, logger *zap.Logger) worker.Worker
}
