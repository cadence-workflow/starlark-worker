package worker

import "github.com/cadence-workflow/starlark-worker/internal"

type (
	RegisterWorkflowOptions = internal.RegisterWorkflowOptions

	RegisterActivityOptions = internal.RegisterActivityOptions
)

type Registry interface {
	RegisterWorkflow(w interface{}, name string)
	RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions)
	RegisterActivity(a interface{})
	RegisterActivityWithOptions(a interface{}, options RegisterActivityOptions)
}

type Worker interface {
	Registry
	Start() error
	// Run is a blocking start and cleans up resources when killed
	// returns error only if it fails to start the worker
	Run(interruptCh <-chan interface{}) error
	// Stop cleans up any resources opened by worker
	Stop()
}
