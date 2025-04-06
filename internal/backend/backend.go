package backend

import (
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.uber.org/zap"
)

type Backend interface {
	RegisterWorker(url string, domain string, taskList string, logger *zap.Logger) worker.Worker
	RegisterWorkflow() workflow.Workflow
}
