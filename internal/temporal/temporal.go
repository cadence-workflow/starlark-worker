package temporal

import (
	"github.com/cadence-workflow/starlark-worker/internal/backend"
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"github.com/uber-go/tally"
	tempactivity "go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	tmpworker "go.temporal.io/sdk/worker"
	temp "go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
	"time"
)

func GetBackend() backend.Backend {
	return temporalBackend{}
}

type temporalBackend struct{}

func (c temporalBackend) RegisterWorkflow() workflow.Workflow {
	return &temporalWorkflow{}
}

func (c temporalBackend) RegisterWorker(url string, domain string, taskList string, logger *zap.Logger) worker.Worker {
	client, err := newClient(url, domain)
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

func (f *temporalFuture) Get(ctx workflow.Context, valPtr interface{}) error {
	return f.f.Get(ctx.(temp.Context), valPtr)
}
func (f *temporalFuture) IsReady() bool {
	return f.f.IsReady()
}

func (worker *temporalWorker) RegisterWorkflow(w interface{}) {
	worker.w.RegisterWorkflow(w)
}

func (worker *temporalWorker) RegisterActivity(a interface{}) {
	worker.w.RegisterActivity(a)
}

func (worker *temporalWorker) Start() error {
	return worker.w.Start()
}

func (worker *temporalWorker) Run(interruptCh <-chan interface{}) error {
	return worker.w.Run(interruptCh)
}

func (worker *temporalWorker) Stop() {
	worker.w.Stop()
}

var _ worker.Worker = (*temporalWorker)(nil)

func (worker *temporalWorker) RegisterWorkflowWithOptions(w interface{}, options worker.RegisterWorkflowOptions) {
	worker.w.RegisterWorkflowWithOptions(w, temp.RegisterOptions{
		Name: options.Name,
		// Optional: Provides a Versioning Behavior to workflows of this type. It is required
		// when WorkerOptions does not specify [DeploymentOptions.DefaultVersioningBehavior],
		// [DeploymentOptions.DeploymentSeriesName] is set, and [UseBuildIDForVersioning] is true.
		// NOTE: Experimental
		VersioningBehavior:            temp.VersioningBehavior(options.VersioningBehavior),
		DisableAlreadyRegisteredCheck: options.DisableAlreadyRegisteredCheck,
	})
}

func (worker *temporalWorker) RegisterActivityWithOptions(w interface{}, options worker.RegisterActivityOptions) {
	worker.w.RegisterActivityWithOptions(w, tempactivity.RegisterOptions{
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

func (w temporalWorkflow) ExecuteActivity(ctx workflow.Context, name string, args ...any) workflow.Future {
	f := temp.ExecuteActivity(ctx.(temp.Context), name, args...)
	return &temporalFuture{f: f}
}

func (w temporalWorkflow) ExecuteChildWorkflow(ctx workflow.Context, name string, args ...any) workflow.Future {
	f := temp.ExecuteChildWorkflow(ctx.(temp.Context), name, args...)
	return &temporalFuture{f: f}
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
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) Now(ctx workflow.Context) time.Time {
	//TODO implement me
	panic("implement me")
}

func (w temporalWorkflow) Sleep(ctx workflow.Context, d time.Duration) (err error) {
	//TODO implement me
	panic("implement me")
}

func newClient(location string, namespace string) (client.Client, error) {
	options := client.Options{
		HostPort:  location,
		Namespace: namespace,
	}

	// Use NewLazyClient to create a lazy-initialized client
	return client.NewLazyClient(options)
}
