package cadence

import (
	"context"
	"github.com/cadence-workflow/starlark-worker/internal"
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"github.com/uber-go/tally"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	cadworker "go.uber.org/cadence/worker"
	cad "go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
	"log"
	"net/url"
	"reflect"
)

type (
	DataConverter = internal.CadenceDataConverter
)

func NewWorkflow() workflow.Workflow {
	return &internal.CadenceWorkflow{}
}

func UpdateWorkflowFunctionContextArgument(w interface{}) interface{} {
	return internal.UpdateWorkflowFunctionContextArgument(w, reflect.TypeOf((*cad.Context)(nil)).Elem())
}

func NewCadenceWorker(url string, domain string, taskList string, logger *zap.Logger) worker.Worker {
	cadInterface := NewInterface(url)
	ctx := context.Background()
	ctx = context.WithValue(ctx, workflow.BackendContextKey, NewWorkflow())
	w := cadworker.New(
		cadInterface,
		domain,
		taskList,
		cadworker.Options{
			MetricsScope: tally.NoopScope,
			Logger:       logger,
			DataConverter: &DataConverter{
				Logger: logger,
			},
			ContextPropagators: []cad.ContextPropagator{
				&HeadersContextPropagator{},
			},
			BackgroundActivityContext: ctx,
		},
	)
	return NewWorker(w)
}

func NewWorker(w cadworker.Worker) worker.Worker {
	return &internal.CadenceWorker{Worker: w}
}

func NewInterface(location string) workflowserviceclient.Interface {
	loc, err := url.Parse(location)
	if err != nil {
		log.Fatalln(err)
	}

	var tran transport.UnaryOutbound
	switch loc.Scheme {
	case "grpc":
		tran = grpc.NewTransport().NewSingleOutbound(loc.Host)
	case "tchannel":
		if t, err := tchannel.NewTransport(tchannel.ServiceName("tchannel")); err != nil {
			log.Fatalln(err)
		} else {
			tran = t.NewSingleOutbound(loc.Host)
		}
	default:
		log.Fatalf("unsupported scheme: %s", loc.Scheme)
	}

	service := "cadence-frontend"
	dispatcher := yarpc.NewDispatcher(yarpc.Config{
		Name: service,
		Outbounds: yarpc.Outbounds{
			service: {
				Unary: tran,
			},
		},
	})
	if err := dispatcher.Start(); err != nil {
		log.Fatalln(err)
	}
	return workflowserviceclient.New(dispatcher.ClientConfig(service))
}
