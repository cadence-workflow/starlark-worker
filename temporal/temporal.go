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
	DataConverter = internal.TemporalDataConverter

	CustomError = internal.TemporalCustomError
)

func NewZapLoggerAdapter(z *zap.Logger) *internal.ZapLoggerAdapter {
	return internal.NewZapLoggerAdapter(z)
}

func NewWorkflow() workflow.Workflow {
	return &internal.TemporalWorkflow{}
}

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

func NewWorker(w tempworker.Worker) worker.Worker {
	return &internal.TemporalWorker{Worker: w}
}

func UpdateWorkflowFunctionContextArgument(w interface{}) (interface{}, string) {
	return internal.UpdateWorkflowFunctionContextArgument(w, reflect.TypeOf((*temp.Context)(nil)).Elem())
}

func NewClient(location string, namespace string) (client.Client, error) {
	scope, closer := tallyv4.NewRootScope(tallyv4.ScopeOptions{
		Prefix: "temporal",
	}, time.Second)
	options := client.Options{
		HostPort:           location,
		Namespace:          namespace,
		DataConverter:      DataConverter{},
		MetricsHandler:     tally.NewMetricsHandler(scope),
		ContextPropagators: []temp.ContextPropagator{&HeadersContextPropagator{}},
	}
	go func() {
		<-time.After(5 * time.Minute) // example shutdown trigger
		closer.Close()
	}()

	// Use NewLazyClient to create a lazy-initialized client
	return client.Dial(options)
}
