package backend

import (
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.uber.org/zap"
)

type Backend interface {
	RegisterWorker(url string, domain string, taskList string, logger *zap.Logger, workerOptions interface{}) worker.Worker
	RegisterWorkflow() workflow.Workflow
}
