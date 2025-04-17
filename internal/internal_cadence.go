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

// CadenceDataConverter is a Cadence encoded.DataConverter that supports Starlark types, such as starlark.String, starlark.Int and others.
type CadenceWorkflow struct{}

// IsCanceledError checks if the error is a CanceledError.
func (w CadenceWorkflow) IsCanceledError(ctx Context, err error) bool {
	var canceledError *cadence.CanceledError
	return errors.As(err, &canceledError)
}

// cadenceWorkflowInfo implements IInfo interface
type cadenceWorkflowInfo struct {
	context cad.Context
}

// cadenceFuture implements Future interface
type cadenceFuture struct {
	f cad.Future
}

// cadenceChildWorkflowFuture implements ChildWorkflowFuture interface
type cadenceChildWorkflowFuture struct {
	cf cad.ChildWorkflowFuture
}

// cadenceSettable implements Settable interface
type cadenceSettable struct {
	s cad.Settable
}

// CadenceWorker implements Worker interface
type CadenceWorker struct {
	Worker cadworker.Worker
}

// RegisterWorkflow registers a workflow with the Cadence worker.
func (w *CadenceWorker) RegisterWorkflow(wf interface{}, funcName string) {
	w.Worker.RegisterWorkflowWithOptions(UpdateWorkflowFunctionContextArgument(wf, reflect.TypeOf((*cad.Context)(nil)).Elem()),
		cad.RegisterOptions{
			Name: funcName,
		})
}

// RegisterActivity registers an activity with the Cadence worker.
func (w *CadenceWorker) RegisterActivity(a interface{}) {
	w.Worker.RegisterActivity(a)
}

// Start starts the Cadence worker.
func (w *CadenceWorker) Start() error {
	return w.Worker.Start()
}

// Run runs the Cadence worker.
func (w *CadenceWorker) Run(_ <-chan interface{}) error {
	return w.Worker.Run()
}

// Stop stops the Cadence worker.
func (w *CadenceWorker) Stop() {
	w.Worker.Stop()
}

// RegisterWorkflowWithOptions registers a workflow with the Cadence worker using options.
func (w *CadenceWorker) RegisterWorkflowWithOptions(runFunc interface{}, options RegisterWorkflowOptions) {
	w.Worker.RegisterWorkflowWithOptions(runFunc, cad.RegisterOptions{
		Name:                          options.Name,
		EnableShortName:               options.EnableShortName,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

// RegisterActivityWithOptions registers an activity with the Cadence worker using options.
func (w *CadenceWorker) RegisterActivityWithOptions(runFunc interface{}, options RegisterActivityOptions) {
	w.Worker.RegisterActivityWithOptions(runFunc, cadactivity.RegisterOptions{
		Name:                          options.Name,
		EnableShortName:               options.EnableShortName,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

// SetValue sets the value of the cadence future.
func (s *cadenceSettable) SetValue(value interface{}) {
	s.s.SetValue(value)
}

// SetError sets the error of the cadence future.
func (s *cadenceSettable) SetError(err error) {
	s.s.SetError(err)
}

// Set sets the value and error of the cadence future.
func (s *cadenceSettable) Set(value interface{}, err error) {
	s.s.Set(value, err)
}

// Chain chains the cadence future with another future.
func (s *cadenceSettable) Chain(future Future) {
	s.s.Chain(future.(*cadenceFuture).f)
}

// Get gets the value of the cadence future.
func (f *cadenceFuture) Get(ctx Context, valPtr interface{}) error {
	return f.f.Get(ctx.(cad.Context), valPtr)
}

// IsReady checks if the cadence future is ready.
func (f *cadenceFuture) IsReady() bool {
	return f.f.IsReady()
}

// Get gets the value of the cadence future.
func (f *cadenceChildWorkflowFuture) Get(ctx Context, valPtr interface{}) error {
	return f.cf.Get(ctx.(cad.Context), valPtr)
}

// IsReady checks if the cadence child workflow future is ready.
func (f *cadenceChildWorkflowFuture) IsReady() bool {
	return f.cf.IsReady()
}

// GetChildWorkflowExecution returns a future that will be ready when child workflow execution started.
func (f *cadenceChildWorkflowFuture) GetChildWorkflowExecution() Future {
	future := f.cf.GetChildWorkflowExecution()
	return &cadenceFuture{f: future}
}

// SignalChildWorkflow sends a signal to the child workflow.
func (f *cadenceChildWorkflowFuture) SignalChildWorkflow(ctx Context, signalName string, data interface{}) Future {
	future := f.cf.SignalChildWorkflow(ctx.(cad.Context), signalName, data)
	return &cadenceFuture{f: future}
}

// ExecutionID returns the execution ID of the workflow.
func (w *cadenceWorkflowInfo) ExecutionID() string {
	return cad.GetInfo(w.context).WorkflowExecution.ID
}

// RunID returns the run ID of the workflow.
func (w *cadenceWorkflowInfo) RunID() string {
	return cad.GetInfo(w.context).WorkflowExecution.RunID
}

// This checks if CadenceWorkflow implements Workflow interface
var _ Workflow = (*CadenceWorkflow)(nil)

// GetLogger is implemented in the Workflow interface to return the logger for the Cadence workflow.
func (w CadenceWorkflow) GetLogger(ctx Context) *zap.Logger {
	return cad.GetLogger(ctx.(cad.Context))
}

// GetActivityLogger is implemented in the Workflow interface to return the logger for the Cadence activity.
func (w CadenceWorkflow) GetActivityLogger(ctx context.Context) *zap.Logger {
	return cadactivity.GetLogger(ctx)
}

// GetActivityInfo returns the activity info for the Cadence workflow.
func (w CadenceWorkflow) GetInfo(ctx Context) IInfo {
	return &cadenceWorkflowInfo{
		context: ctx.(cad.Context),
	}
}

// ExecuteActivity executes an activity in the Cadence workflow.
func (w CadenceWorkflow) ExecuteActivity(ctx Context, activity interface{}, args ...interface{}) Future {
	f := cad.ExecuteActivity(ctx.(cad.Context), activity, args...)
	return &cadenceFuture{f: f}
}

// WithValue sets a value in the Cadence workflow context.
func (w CadenceWorkflow) WithValue(parent Context, key interface{}, val interface{}) Context {
	return cad.WithValue(parent.(cad.Context), key, val)
}

// NewDisconnectedContext creates a new disconnected context for the Cadence workflow.
func (w CadenceWorkflow) NewDisconnectedContext(parent Context) (ctx Context, cancel func()) {
	return cad.NewDisconnectedContext(parent.(cad.Context))
}

// GetMetricsScope returns the metrics scope for the Cadence workflow.
func (w CadenceWorkflow) GetMetricsScope(ctx Context) interface{} {
	return cad.GetMetricsScope(ctx.(cad.Context))
}

// WithTaskList sets the task list for the Cadence workflow context.
func (w *CadenceWorkflow) WithTaskList(ctx Context, name string) Context {
	return cad.WithTaskList(ctx.(cad.Context), name)
}

// WithActivityOptions sets the activity options for the Cadence workflow context.
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

// WithChildOptions sets the child workflow options for the Cadence workflow context.
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

// WithRetryPolicy sets the retry policy for the Cadence workflow context.
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

// SetQueryHandler sets a query handler for the Cadence workflow.
func (w CadenceWorkflow) SetQueryHandler(ctx Context, queryType string, handler interface{}) error {
	return cad.SetQueryHandler(ctx.(cad.Context), queryType, handler)
}

// WithWorkflowDomain sets the workflow domain for the Cadence workflow context.
func (w CadenceWorkflow) WithWorkflowDomain(ctx Context, name string) Context {
	return cad.WithWorkflowDomain(ctx.(cad.Context), name)
}

// WithWorkflowTaskList sets the workflow task list for the Cadence workflow context.
func (w CadenceWorkflow) WithWorkflowTaskList(ctx Context, name string) Context {
	return cad.WithWorkflowTaskList(ctx.(cad.Context), name)
}

// ExecuteChildWorkflow executes a child workflow in the Cadence workflow.
func (w CadenceWorkflow) ExecuteChildWorkflow(ctx Context, childWorkflow interface{}, args ...interface{}) ChildWorkflowFuture {
	f := cad.ExecuteChildWorkflow(ctx.(cad.Context), childWorkflow, args...)
	return &cadenceChildWorkflowFuture{cf: f}
}

// NewCustomError creates a new custom error for the Cadence workflow.
func (w CadenceWorkflow) NewCustomError(reason string, details ...interface{}) CustomError {
	return cadence.NewCustomError(reason, details...)
}

// NewFuture creates a new future for the Cadence workflow.
func (w CadenceWorkflow) NewFuture(ctx Context) (Future, Settable) {
	f, s := cad.NewFuture(ctx.(cad.Context))
	return &cadenceFuture{f: f}, &cadenceSettable{s: s}
}

// Go executes a function in the Cadence workflow context.
func (w CadenceWorkflow) Go(ctx Context, f func(ctx Context)) {
	cad.Go(ctx.(cad.Context), func(c cad.Context) {
		f(c)
	})
}

// SideEffect executes a side effect in the Cadence workflow context.
func (w CadenceWorkflow) SideEffect(ctx Context, f func(ctx Context) interface{}) encoded.Value {
	return cad.SideEffect(ctx.(cad.Context), func(c cad.Context) interface{} {
		return f(c)
	})
}

// Now returns the current time in the Cadence workflow context.
func (w CadenceWorkflow) Now(ctx Context) time.Time {
	return cad.Now(ctx.(cad.Context))
}

// Sleep pauses the Cadence workflow for a specified duration.
func (w CadenceWorkflow) Sleep(ctx Context, d time.Duration) (err error) {
	return cad.Sleep(ctx.(cad.Context), d)
}

// NewInterface creates a new Cadence workflow service client interface.
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

// ToData encodes the given values into a byte slice.
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

// FromData decodes the given byte slice into the specified values.
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
