package backend

import (
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.uber.org/zap"
)

type Backend interface {
	RegisterWorkflow() workflow.Workflow
	RegisterWorker(url string, domain string, taskList string, logger *zap.Logger) worker.Worker
}
