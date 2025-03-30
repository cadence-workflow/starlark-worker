package temporal

import (
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"github.com/uber-go/tally"
	"go.temporal.io/sdk/temporal"
	temp "go.temporal.io/sdk/workflow"
	"go.uber.org/zap"
)

type temporalWorkflow struct{}

type workflowInfo struct {
	context temp.Context
}

type temporalFuture struct {
	f temp.Future
}

func (f *temporalFuture) Get(ctx workflow.Context, valPtr interface{}) error {
	return f.f.Get(ctx.(temp.Context), valPtr)
}

func (f *temporalFuture) IsReady() bool {
	return f.f.IsReady()
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
