package worker

type Worker interface {
	Registry
	Start() error
	// Run is a blocking start and cleans up resources when killed
	// returns error only if it fails to start the worker
	Run() error
	// Stop cleans up any resources opened by worker
	Stop()
}

type Registry interface {
	RegisterWorkflow(w interface{})
	RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions)
	RegisterActivity(a interface{})
	RegisterActivityWithOptions(a interface{}, options RegisterActivityOptions)
}

var backendWorker Registry

func RegisterBackend(b Registry) {
	backendWorker = b
}

type RegisterWorkflowOptions struct {
	Name string
	// Workflow type name is equal to function name instead of fully qualified name including function package.
	// This option has no effect when explicit Name is provided.
	EnableShortName               bool
	DisableAlreadyRegisteredCheck bool
}

func RegisterWorkflow(w interface{}) {
	if backendWorker == nil {
		panic("workflow backendWorker not registered")
	}
	backendWorker.RegisterWorkflow(w)
}

// RegisterWorkflowWithOptions registers the workflow function with options.
// The user can use options to provide an external name for the workflow or leave it empty if no
// external name is required. This can be used as
//
//	worker.RegisterWorkflowWithOptions(sampleWorkflow, RegisterWorkflowOptions{})
//	worker.RegisterWorkflowWithOptions(sampleWorkflow, RegisterWorkflowOptions{Name: "foo"})
//
// This method panics if workflowFunc doesn't comply with the expected format or tries to register the same workflow
// type name twice. Use workflow.RegisterOptions.DisableAlreadyRegisteredCheck to allow multiple registrations.
func RegisterWorkflowWithOptions(w interface{}, options RegisterWorkflowOptions) {
	if backendWorker == nil {
		panic("workflow backendWorker not registered")
	}
	backendWorker.RegisterWorkflowWithOptions(w, options)
}

func RegisterActivity(a interface{}) {
	if backendWorker == nil {
		panic("workflow backendWorker not registered")
	}
	backendWorker.RegisterActivity(a)
}
func RegisterActivityWithOptions(a interface{}, options RegisterActivityOptions) {
	if backendWorker == nil {
		panic("workflow backendWorker not registered")
	}
	backendWorker.RegisterActivityWithOptions(a, options)
}

type RegisterActivityOptions struct {
	// When an activity is a function the name is an actual activity type name.
	// When an activity is part of a structure then each member of the structure becomes an activity with
	// this Name as a prefix + activity function name.
	Name string
	// Activity type name is equal to function name instead of fully qualified
	// name including function package (and struct type if used).
	// This option has no effect when explicit Name is provided.
	EnableShortName               bool
	DisableAlreadyRegisteredCheck bool
	// Automatically send heartbeats for this activity at an interval that is less than the HeartbeatTimeout.
	// This option has no effect if the activity is executed with a HeartbeatTimeout of 0.
	// Default: false
	EnableAutoHeartbeat bool
}
