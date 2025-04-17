package temporal

import (
	"context"
	"github.com/cadence-workflow/starlark-worker/internal"
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	tallyv4 "github.com/uber-go/tally/v4"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/contrib/tally"
	tempworker "go.temporal.io/sdk/worker"
	temp "go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"reflect"
	"time"
)

type (
	// DataConverter is a Temporal-specific implementation that encodes and decodes
	// arguments/results passed between workflows and activities.
	//
	// It wraps JSON serialization but may add custom logic such as compression or logging.
	//
	// Used in: NewClient (via client.Options.DataConverter)
	DataConverter = internal.TemporalDataConverter

	// CustomError represents a structured application error that is serializable across
	// Temporal boundaries. Allows passing "reason" and details in workflow/activity errors.
	//
	// Example:
	//   return temporal.CustomError{Reason: "failed_validation", Details: "missing field"}
	CustomError = internal.TemporalCustomError
)

// NewZapLoggerAdapter converts a zap.Logger into a Temporal-compatible logger.
// This is useful because Temporal expects its own logging interface, and this bridges that gap.
//
// Example:
//
//	temporalLogger := temporal.NewZapLoggerAdapter(zapLogger)
//	client.NewClient(..., client.Options{Logger: temporalLogger})
func NewZapLoggerAdapter(z *zap.Logger) *internal.ZapLoggerAdapter {
	return internal.NewZapLoggerAdapter(z)
}

// NewWorkflow returns a new instance of TemporalWorkflow, which implements the shared
// `workflow.Workflow` interface. This enables pluggability across Cadence/Temporal.
//
// This instance is injected into the workflow.Context under the `BackendContextKey`
// so that backend-specific behavior can be triggered dynamically.
func NewWorkflow() workflow.Workflow {
	return &internal.TemporalWorkflow{}
}

// NewTemporalWorker creates a new Temporal worker given a URL, namespace, and task queue.
// It connects to the Temporal service, prepares a base context with backend set, and initializes the worker.
//
// This is the standard entrypoint for spinning up a Temporal-based worker for executing workflows.
func NewTemporalWorker(url string, namespace string, taskQueue string) worker.Worker {
	newClient, err := NewClient(url, namespace)
	if err != nil {
		panic("failed to create temporal client")
	}

	ctx := context.Background()
	ctx = context.WithValue(ctx, workflow.BackendContextKey, NewWorkflow())

	w := tempworker.New(
		newClient,
		taskQueue,
		tempworker.Options{
			BackgroundActivityContext: ctx,
		},
	)
	return NewWorker(w)
}

// NewWorker wraps the Temporal worker in a common `worker.Worker` interface,
// which abstracts over Cadence or Temporal backends.
//
// Used when you want to run the worker inside the starlark runtime or with unified tooling.
func NewWorker(w tempworker.Worker) worker.Worker {
	return &internal.TemporalWorker{Worker: w}
}

// UpdateWorkflowFunctionContextArgument converts a workflow function's signature to conform
// to the expected Temporal format by swapping in a Temporal-specific context.
//
// This function supports dynamic introspection and adaptation of user-defined workflows.
func UpdateWorkflowFunctionContextArgument(w interface{}) interface{} {
	return internal.UpdateWorkflowFunctionContextArgument(w, reflect.TypeOf((*temp.Context)(nil)).Elem())
}

// NewClient creates a new Temporal SDK client with configuration like:
// - HostPort
// - Namespace
// - Custom DataConverter
// - Metrics via Tally
// - Custom context propagators
//
// This function is typically called once when spinning up a worker.
//
// It also starts a background goroutine that auto-closes the metrics scope after 5 minutes.
func NewClient(location string, namespace string) (client.Client, error) {
	// TODO handler close in worker lifecycle management method in main.go
	scope, _ := tallyv4.NewRootScope(tallyv4.ScopeOptions{
		Prefix: "temporal",
	}, time.Second)
	options := client.Options{
		HostPort:           location,
		Namespace:          namespace,
		DataConverter:      DataConverter{},
		MetricsHandler:     tally.NewMetricsHandler(scope),
		ContextPropagators: []temp.ContextPropagator{&HeadersContextPropagator{}},
	}
	return client.Dial(options)
}
