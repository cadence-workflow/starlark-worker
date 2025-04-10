package internal

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"github.com/cadence-workflow/starlark-worker/encoded"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/star"
	jsoniter "github.com/json-iterator/go"
	"go.starlark.net/starlark"
	"go.uber.org/cadence"
	"go.uber.org/cadence/.gen/go/cadence/workflowserviceclient"
	cadactivity "go.uber.org/cadence/activity"
	cadworker "go.uber.org/cadence/worker"
	cad "go.uber.org/cadence/workflow"
	"go.uber.org/yarpc"
	"go.uber.org/yarpc/api/transport"
	"go.uber.org/yarpc/transport/grpc"
	"go.uber.org/yarpc/transport/tchannel"
	"go.uber.org/zap"
	"io"
	"log"
	"net/url"
	"reflect"
	"time"
)

type CadenceWorkflow struct{}

func (w CadenceWorkflow) IsCanceledError(ctx Context, err error) bool {
	var canceledError *cadence.CanceledError
	return errors.As(err, &canceledError)
}

type cadenceWorkflowInfo struct {
	context cad.Context
}

type cadenceFuture struct {
	f cad.Future
}

type cadenceChildWorkflowFuture struct {
	cf cad.ChildWorkflowFuture
}

type cadenceSettable struct {
	s cad.Settable
}

type CadenceWorker struct {
	Worker cadworker.Worker
}

func (w *CadenceWorker) RegisterWorkflow(wf interface{}) {
	w.Worker.RegisterWorkflow(UpdateWorkflowFunctionContextArgument(wf, reflect.TypeOf((*cad.Context)(nil)).Elem()))
}

func (w *CadenceWorker) RegisterActivity(a interface{}) {
	w.Worker.RegisterActivity(a)
}

func (w *CadenceWorker) Start() error {
	return w.Worker.Start()
}

func (w *CadenceWorker) Run(_ <-chan interface{}) error {
	return w.Worker.Run()
}

func (w *CadenceWorker) Stop() {
	w.Worker.Stop()
}

func (w *CadenceWorker) RegisterWorkflowWithOptions(runFunc interface{}, options RegisterWorkflowOptions) {
	w.Worker.RegisterWorkflowWithOptions(runFunc, cad.RegisterOptions{
		Name:                          options.Name,
		EnableShortName:               options.EnableShortName,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

func (w *CadenceWorker) RegisterActivityWithOptions(runFunc interface{}, options RegisterActivityOptions) {
	w.Worker.RegisterActivityWithOptions(runFunc, cadactivity.RegisterOptions{
		Name:                          options.Name,
		EnableShortName:               options.EnableShortName,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

func (s *cadenceSettable) SetValue(value interface{}) {
	s.s.SetValue(value)
}

func (s *cadenceSettable) SetError(err error) {
	s.s.SetError(err)
}

func (s *cadenceSettable) Set(value interface{}, err error) {
	s.s.Set(value, err)
}

func (s *cadenceSettable) Chain(future Future) {
	s.s.Chain(future.(*cadenceFuture).f)
}

func (f *cadenceFuture) Get(ctx Context, valPtr interface{}) error {
	return f.f.Get(ctx.(cad.Context), valPtr)
}

func (f *cadenceFuture) IsReady() bool {
	return f.f.IsReady()
}

func (f *cadenceChildWorkflowFuture) Get(ctx Context, valPtr interface{}) error {
	return f.cf.Get(ctx.(cad.Context), valPtr)
}

func (f *cadenceChildWorkflowFuture) IsReady() bool {
	return f.cf.IsReady()
}
func (f *cadenceChildWorkflowFuture) GetChildWorkflowExecution() Future {
	future := f.cf.GetChildWorkflowExecution()
	return &cadenceFuture{f: future}
}

func (f *cadenceChildWorkflowFuture) SignalChildWorkflow(ctx Context, signalName string, data interface{}) Future {
	future := f.cf.SignalChildWorkflow(ctx.(cad.Context), signalName, data)
	return &cadenceFuture{f: future}
}

func (w *cadenceWorkflowInfo) ExecutionID() string {
	return cad.GetInfo(w.context).WorkflowExecution.ID
}
func (w *cadenceWorkflowInfo) RunID() string {
	return cad.GetInfo(w.context).WorkflowExecution.RunID
}

var _ Workflow = (*CadenceWorkflow)(nil)

func (w CadenceWorkflow) GetLogger(ctx Context) *zap.Logger {
	return cad.GetLogger(ctx.(cad.Context))
}

func (w CadenceWorkflow) GetActivityLogger(ctx context.Context) *zap.Logger {
	return cadactivity.GetLogger(ctx)
}

func (w CadenceWorkflow) GetInfo(ctx Context) IInfo {
	return &cadenceWorkflowInfo{
		context: ctx.(cad.Context),
	}
}

func (w CadenceWorkflow) ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	f := cad.ExecuteActivity(ctx.(cad.Context), activity, args...)
	return &cadenceFuture{f: f}
}

func (w CadenceWorkflow) WithValue(parent Context, key interface{}, val interface{}) Context {
	return cad.WithValue(parent.(cad.Context), key, val)
}

func (w CadenceWorkflow) NewDisconnectedContext(parent Context) (ctx Context, cancel func()) {
	return cad.NewDisconnectedContext(parent.(cad.Context))
}

func (w CadenceWorkflow) GetMetricsScope(ctx Context) interface{} {
	return cad.GetMetricsScope(ctx.(cad.Context))
}

func (w *CadenceWorkflow) WithTaskList(ctx Context, name string) Context {
	return cad.WithTaskList(ctx.(cad.Context), name)
}

func (w *CadenceWorkflow) WithActivityOptions(ctx Context, options ActivityOptions) Context {
	cadOptions := cad.ActivityOptions{
		TaskList:               options.TaskList,
		ScheduleToCloseTimeout: options.ScheduleToCloseTimeout,
		ScheduleToStartTimeout: options.ScheduleToStartTimeout,
		StartToCloseTimeout:    options.StartToCloseTimeout,
		HeartbeatTimeout:       options.HeartbeatTimeout,
		WaitForCancellation:    options.WaitForCancellation,
		ActivityID:             options.ActivityID,
	}
	if options.RetryPolicy != nil {
		cadOptions.RetryPolicy = &cad.RetryPolicy{
			InitialInterval:          options.RetryPolicy.InitialInterval,
			BackoffCoefficient:       options.RetryPolicy.BackoffCoefficient,
			MaximumInterval:          options.RetryPolicy.MaximumInterval,
			ExpirationInterval:       options.RetryPolicy.ExpirationInterval,
			MaximumAttempts:          options.RetryPolicy.MaximumAttempts,
			NonRetriableErrorReasons: options.RetryPolicy.NonRetriableErrorReasons,
		}
	}
	return cad.WithActivityOptions(ctx.(cad.Context), cadOptions)
}

func (w CadenceWorkflow) WithChildOptions(ctx Context, cwo ChildWorkflowOptions) Context {
	cadOptions := cad.ChildWorkflowOptions{
		Domain:                       cwo.Domain,
		WorkflowID:                   cwo.WorkflowID,
		TaskList:                     cwo.TaskList,
		ExecutionStartToCloseTimeout: cwo.ExecutionStartToCloseTimeout,
		TaskStartToCloseTimeout:      cwo.TaskStartToCloseTimeout,
		WaitForCancellation:          cwo.WaitForCancellation,
		CronSchedule:                 cwo.CronSchedule,
		Memo:                         cwo.Memo,
		SearchAttributes:             cwo.SearchAttributes,
	}
	if cwo.RetryPolicy != nil {
		cadOptions.RetryPolicy = &cad.RetryPolicy{
			InitialInterval:          cwo.RetryPolicy.InitialInterval,
			BackoffCoefficient:       cwo.RetryPolicy.BackoffCoefficient,
			MaximumInterval:          cwo.RetryPolicy.MaximumInterval,
			ExpirationInterval:       cwo.RetryPolicy.ExpirationInterval,
			MaximumAttempts:          cwo.RetryPolicy.MaximumAttempts,
			NonRetriableErrorReasons: cwo.RetryPolicy.NonRetriableErrorReasons,
		}
	}
	return cad.WithChildOptions(ctx.(cad.Context), cadOptions)
}

func (w CadenceWorkflow) WithRetryPolicy(ctx Context, retryPolicy RetryPolicy) Context {
	cadRetryPolicy := cad.RetryPolicy{
		InitialInterval:          retryPolicy.InitialInterval,
		BackoffCoefficient:       retryPolicy.BackoffCoefficient,
		MaximumInterval:          retryPolicy.MaximumInterval,
		ExpirationInterval:       retryPolicy.ExpirationInterval,
		MaximumAttempts:          retryPolicy.MaximumAttempts,
		NonRetriableErrorReasons: retryPolicy.NonRetriableErrorReasons,
	}
	return cad.WithRetryPolicy(ctx.(cad.Context), cadRetryPolicy)
}

func (w CadenceWorkflow) SetQueryHandler(ctx Context, queryType string, handler interface{}) error {
	return cad.SetQueryHandler(ctx.(cad.Context), queryType, handler)
}

func (w CadenceWorkflow) WithWorkflowDomain(ctx Context, name string) Context {
	return cad.WithWorkflowDomain(ctx.(cad.Context), name)
}

func (w CadenceWorkflow) WithWorkflowTaskList(ctx Context, name string) Context {
	return cad.WithWorkflowTaskList(ctx.(cad.Context), name)
}

func (w CadenceWorkflow) ExecuteChildWorkflow(ctx Context, childWorkflow interface{}, args ...interface{}) ChildWorkflowFuture {
	f := cad.ExecuteChildWorkflow(ctx.(cad.Context), childWorkflow, args...)
	return &cadenceChildWorkflowFuture{cf: f}
}

func (w CadenceWorkflow) NewCustomError(reason string, details ...interface{}) CustomError {
	return cadence.NewCustomError(reason, details...)
}

func (w CadenceWorkflow) NewFuture(ctx Context) (Future, Settable) {
	f, s := cad.NewFuture(ctx.(cad.Context))
	return &cadenceFuture{f: f}, &cadenceSettable{s: s}
}

func (w CadenceWorkflow) Go(ctx Context, f func(ctx Context)) {
	cad.Go(ctx.(cad.Context), func(c cad.Context) {
		f(c)
	})
}

func (w CadenceWorkflow) SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value {
	return cad.SideEffect(ctx.(cad.Context), func(c cad.Context) interface{} {
		return f(c)
	})
}

func (w CadenceWorkflow) Now(ctx Context) time.Time {
	return cad.Now(ctx.(cad.Context))
}

func (w CadenceWorkflow) Sleep(ctx Context, d time.Duration) (err error) {
	return cad.Sleep(ctx.(cad.Context), d)
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

const cadenceDelimiter byte = '\n'

// CadenceDataConverter is a Cadence encoded.DataConverter that supports Starlark types, such as starlark.String, starlark.Int and others.
// Enables passing Starlark values between Cadence workflows and activities.
type CadenceDataConverter struct {
	Logger *zap.Logger
}

func (s *CadenceDataConverter) ToData(values ...any) ([]byte, error) {
	if len(values) == 1 {
		switch v := values[0].(type) {
		case []byte:
			return v, nil
		case starlark.Bytes:
			return []byte(v), nil
		}
	}
	var buf bytes.Buffer
	for _, v := range values {
		b, err := star.Encode(v) // try star encoder
		if _, ok := err.(star.UnsupportedTypeError); ok {
			b, err = jsoniter.Marshal(v) // go encoder fallback
		}
		if err != nil {
			s.Logger.Error("encode-error", ext.ZapError(err)...)
			return nil, err
		}
		buf.Write(b)
		buf.WriteByte(cadenceDelimiter)
	}
	return buf.Bytes(), nil
}

func (s *CadenceDataConverter) FromData(data []byte, to ...any) error {
	if len(to) == 1 {
		switch to := to[0].(type) {
		case *[]byte:
			*to = data
			return nil
		case *starlark.Bytes:
			*to = starlark.Bytes(data)
			return nil
		}
	}
	r := bufio.NewReader(bytes.NewReader(data))

	for i := 0; ; i++ {
		line, err := r.ReadBytes(cadenceDelimiter)
		var eof bool
		if err == io.EOF {
			eof = true
			err = nil
		}
		if err != nil {
			s.Logger.Error("decode-error", ext.ZapError(err)...)
			return err
		}
		ll := len(line)
		if eof && ll == 0 {
			break
		}
		if line[ll-1] == cadenceDelimiter {
			line = line[:ll-1]
		}
		out := to[i]

		// First try to decode the value using Starlark decoder, and if it fails, fall back to JSON decoder.
		err = star.Decode(line, out)
		if _, ok := err.(star.UnsupportedTypeError); ok {
			err = jsoniter.Unmarshal(line, out)
		}
		if err != nil {
			s.Logger.Error("decode-error", ext.ZapError(err)...)
			return err
		}
		if eof {
			break
		}
	}
	return nil
}
