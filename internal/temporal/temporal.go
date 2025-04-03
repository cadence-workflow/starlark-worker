package temporal

import (
	"github.com/cadence-workflow/starlark-worker/internal/backend"
	"github.com/cadence-workflow/starlark-worker/internal/encoded"
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"github.com/uber-go/tally"
	tempactivity "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	tmpworker "go.temporal.io/sdk/worker"
	temp "go.temporal.io/sdk/workflow"
	cad "go.uber.org/cadence/workflow"
	"go.uber.org/zap"
	"reflect"
	"time"
)

func GetBackend() backend.Backend {
	return temporalBackend{}
}

var _ backend.Backend = (*temporalBackend)(nil)

type temporalBackend struct{}

func (c temporalBackend) RegisterWorker(url string, domain string, taskList string, logger *zap.Logger) worker.Worker {
	client, err := NewClient(url, domain)
	if err != nil {
		panic("failed to create temporal client")
	}
	worker := tmpworker.New(
		client,
		taskList,
		tmpworker.Options{},
	)
	return &temporalWorker{w: worker}
}

type temporalWorkflow struct{}

type workflowInfo struct {
	context temp.Context
}

type temporalFuture struct {
	f temp.Future
}

type temporalChildWorkflowFuture struct {
	cf temp.ChildWorkflowFuture
}

type temporalSettable struct {
	s temp.Settable
}

type temporalWorker struct {
	w tmpworker.Worker
}

var _ worker.Worker = (*temporalWorker)(nil)

func (f *temporalFuture) Get(ctx workflow.Context, valPtr interface{}) error {
	return f.f.Get(ctx.(temp.Context), valPtr)
}
func (f *temporalFuture) IsReady() bool {
	return f.f.IsReady()
}

func (tw *temporalWorker) RegisterWorkflow(wf interface{}) {
	originalFunc := reflect.ValueOf(wf)
	originalType := originalFunc.Type()

	if originalType.Kind() != reflect.Func || originalType.NumIn() == 0 {
		panic("workflow function must be a function and have at least one argument (context)")
	}

	// Build a new function with the same signature but context converted to cadence.Context
	wrappedFuncType := reflect.FuncOf(
		append([]reflect.Type{reflect.TypeOf((*cad.Context)(nil)).Elem()}, worker.GetRemainingInTypes(originalType)...),
		worker.GetOutTypes(originalType),
		false,
	)

	wrappedFunc := reflect.MakeFunc(wrappedFuncType, func(args []reflect.Value) []reflect.Value {
		// Replace cadence.Context with workflow.Context for original call
		newArgs := make([]reflect.Value, len(args))
		newArgs[0] = args[0].Convert(reflect.TypeOf((*temp.Context)(nil)).Elem()) // keep as cadence.Context
		for i := 1; i < len(args); i++ {
			newArgs[i] = args[i]
		}
		return originalFunc.Call(newArgs)
	})

	tw.w.RegisterWorkflow(wrappedFunc.Interface())
}
func (tw *temporalWorker) RegisterActivity(a interface{}) {
	tw.w.RegisterActivity(a)
}

func (tw *temporalWorker) Start() error {
	return tw.w.Start()
}

func (tw *temporalWorker) Stop() {
	tw.w.Stop()
}

func (tw *temporalWorker) Run(interruptCh <-chan interface{}) error {
	return tw.w.Run(interruptCh)
}

func (tw *temporalWorker) RegisterWorkflowWithOptions(w interface{}, options worker.RegisterWorkflowOptions) {
	tw.w.RegisterWorkflowWithOptions(w, temp.RegisterOptions{
		Name: options.Name,
		// Optional: Provides a Versioning Behavior to workflows of this type. It is required
		// when WorkerOptions does not specify [DeploymentOptions.DefaultVersioningBehavior],
		// [DeploymentOptions.DeploymentSeriesName] is set, and [UseBuildIDForVersioning] is true.
		// NOTE: Experimental
		VersioningBehavior:            temp.VersioningBehavior(options.VersioningBehavior),
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

func (tw *temporalWorker) RegisterActivityWithOptions(w interface{}, options worker.RegisterActivityOptions) {
	tw.w.RegisterActivityWithOptions(w, tempactivity.RegisterOptions{
		Name:                          options.Name,
		SkipInvalidStructFunctions:    options.SkipInvalidStructFunctions,
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

func (s *temporalSettable) SetValue(value interface{}) {
	s.s.SetValue(value)
}

func (s *temporalSettable) SetError(err error) {
	s.s.SetError(err)
}

func (s *temporalSettable) Set(value interface{}, err error) {
	s.s.Set(value, err)
}

func (s *temporalSettable) Chain(future workflow.Future) {
	s.s.Chain(future.(*temporalFuture).f)
}

func (f *temporalChildWorkflowFuture) Get(ctx workflow.Context, valPtr interface{}) error {
	return f.cf.Get(ctx.(temp.Context), valPtr)
}

func (f *temporalChildWorkflowFuture) IsReady() bool {
	return f.cf.IsReady()
}
func (f *temporalChildWorkflowFuture) GetChildWorkflowExecution() workflow.Future {
	future := f.cf.GetChildWorkflowExecution()
	return &temporalFuture{f: future}
}

func (f *temporalChildWorkflowFuture) SignalChildWorkflow(ctx workflow.Context, signalName string, data interface{}) workflow.Future {
	future := f.cf.SignalChildWorkflow(ctx.(temp.Context), signalName, data)
	return &temporalFuture{f: future}
}

func (w *workflowInfo) ExecutionID() string {
	return temp.GetInfo(w.context).WorkflowExecution.ID
}
func (w *workflowInfo) RunID() string {
	return temp.GetInfo(w.context).WorkflowExecution.RunID
}

var _ workflow.Workflow = (*temporalWorkflow)(nil)

func (w temporalWorkflow) GetLogger(ctx workflow.Context) *zap.Logger {
	logger := temp.GetLogger(ctx.(temp.Context))
	if zl, ok := logger.(*ZapLoggerAdapter); ok {
		zap := zl.Zap()
		return zap
	}
	return nil
}

func (w temporalWorkflow) GetInfo(ctx workflow.Context) workflow.IInfo {
	return &workflowInfo{
		context: ctx.(temp.Context),
	}
}

func (w temporalWorkflow) ExecuteActivity(ctx workflow.Context, activity interface{}, args ...interface{}) workflow.Future {
	f := temp.ExecuteActivity(ctx.(temp.Context), activity, args...)
	return &temporalFuture{f: f}
}

func (w temporalWorkflow) ExecuteChildWorkflow(ctx workflow.Context, name interface{}, args ...interface{}) workflow.ChildWorkflowFuture {
	f := temp.ExecuteChildWorkflow(ctx.(temp.Context), name, args...)
	return &temporalChildWorkflowFuture{cf: f}
}

func (w temporalWorkflow) WithValue(parent workflow.Context, key interface{}, val interface{}) workflow.Context {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) NewDisconnectedContext(parent workflow.Context) (ctx workflow.Context, cancel func()) {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) GetMetricsScope(ctx workflow.Context) tally.Scope {
	//TODO implement me
	panic("implement me")
}

func (w *temporalWorkflow) WithTaskList(ctx workflow.Context, name string) workflow.Context {
	//TODO implement me
	panic("implement me")
}

func (w *temporalWorkflow) WithActivityOptions(ctx workflow.Context, options workflow.ActivityOptions) workflow.Context {
	cadOptions := temp.ActivityOptions{
		TaskQueue:              options.TaskList,
		ScheduleToCloseTimeout: options.ScheduleToCloseTimeout,
		ScheduleToStartTimeout: options.ScheduleToStartTimeout,
		StartToCloseTimeout:    options.StartToCloseTimeout,
		HeartbeatTimeout:       options.HeartbeatTimeout,
		WaitForCancellation:    options.WaitForCancellation,
		ActivityID:             options.ActivityID,
	}
	if options.RetryPolicy != nil {
		cadOptions.RetryPolicy = &temporal.RetryPolicy{
			NonRetryableErrorTypes: options.RetryPolicy.NonRetriableErrorReasons,
			InitialInterval:        options.RetryPolicy.InitialInterval,
			BackoffCoefficient:     options.RetryPolicy.BackoffCoefficient,
			MaximumInterval:        options.RetryPolicy.MaximumInterval,
			MaximumAttempts:        options.RetryPolicy.MaximumAttempts,
		}
	}
	return temp.WithActivityOptions(ctx.(temp.Context), cadOptions)
}

func (w temporalWorkflow) WithChildOptions(ctx workflow.Context, cwo workflow.ChildWorkflowOptions) workflow.Context {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) SetQueryHandler(ctx workflow.Context, queryType string, handler interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) WithWorkflowDomain(ctx workflow.Context, name string) workflow.Context {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) WithWorkflowTaskList(ctx workflow.Context, name string) workflow.Context {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) NewCustomError(reason string, details ...interface{}) error {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) NewFuture(ctx workflow.Context) (workflow.Future, workflow.Settable) {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) SideEffect(ctx workflow.Context, f func(ctx workflow.Context) interface{}) encoded.Value {
	return temp.SideEffect(ctx.(temp.Context), func(ctx temp.Context) interface{} {
		return f(ctx)
	})
}

func (w temporalWorkflow) Now(ctx workflow.Context) time.Time {
	return temp.Now(ctx.(temp.Context))
}

func (w temporalWorkflow) Sleep(ctx workflow.Context, d time.Duration) (err error) {
	return temp.Sleep(ctx.(temp.Context), d)
}

func (w temporalWorkflow) Go(ctx workflow.Context, f func(ctx workflow.Context)) {
	temp.Go(ctx.(temp.Context), func(ctx temp.Context) {
		f(ctx)
	})
}

func NewClient(location string, namespace string) (client.Client, error) {
	options := client.Options{
		HostPort:      location,
		Namespace:     namespace,
		DataConverter: DataConverter{},
	}

	// Use NewLazyClient to create a lazy-initialized client
	return client.Dial(options)
}
