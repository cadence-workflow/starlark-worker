package progress

import (
	"fmt"

	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
)

type Module struct{}

var _ starlark.HasAttrs = &Module{}

func (f *Module) String() string                        { return pluginID }
func (f *Module) Type() string                          { return pluginID }
func (f *Module) Freeze()                               {}
func (f *Module) Truth() starlark.Bool                  { return true }
func (f *Module) Hash() (uint32, error)                 { return 0, fmt.Errorf("no-hash") }
func (f *Module) Attr(n string) (starlark.Value, error) { return star.Attr(f, n, builtins, properties) }
func (f *Module) AttrNames() []string                   { return star.AttrNames(builtins, properties) }

const (
	TaskProgressQueryHandlerKey = "task_progress"
	TaskStatePending            = "PENDING"
	TaskStateRunning            = "RUNNING"
	TaskStateSucceeded          = "SUCCEEDED"
	TaskStateFailed             = "FAILED"
	TaskStateKilled             = "KILLED"
	TaskStateSkipped            = "SKIPPED"
)

var builtins = map[string]*starlark.Builtin{
	"report": starlark.NewBuiltin("report", report),
}

var properties = map[string]star.PropertyFactory{
	"task_state_running":   _getRunningState,
	"task_state_pending":   _getPendingState,
	"task_state_succeeded": _getSucceededState,
	"task_state_failed":    _getFailedState,
	"task_state_killed":    _getKilledState,
	"task_state_skipped":   _getSkippedState,
}

// Report reports a progress string from Go code
func Report(ctx workflow.Context, progressStr string) error {
	logger := workflow.GetLogger(ctx)
	logger.Info(TaskProgressQueryHandlerKey, zap.String("msg", progressStr))
	progress := service.GetProgress(ctx)
	// Note that this push back is not thread-safe
	progress.PushBack(progressStr)
	return nil
}

func report(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	// report(progress: str)
	// Report a progress string

	ctx := service.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var progressStr starlark.String

	if err := starlark.UnpackArgs("report", args, kwargs, TaskProgressQueryHandlerKey, &progressStr); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	logger.Info(TaskProgressQueryHandlerKey, zap.String("msg", string(progressStr)))
	progress := service.GetProgress(ctx)
	// Note that this push back is not thread-safe
	progress.PushBack(string(progressStr))
	return starlark.None, nil
}

func _getPendingState(reciever starlark.Value) (starlark.Value, error) {
	return starlark.String(TaskStatePending), nil
}

func _getRunningState(reciever starlark.Value) (starlark.Value, error) {
	return starlark.String(TaskStateRunning), nil
}

func _getSucceededState(reciever starlark.Value) (starlark.Value, error) {
	return starlark.String(TaskStateSucceeded), nil
}

func _getFailedState(reciever starlark.Value) (starlark.Value, error) {
	return starlark.String(TaskStateFailed), nil
}

func _getKilledState(reciever starlark.Value) (starlark.Value, error) {
	return starlark.String(TaskStateKilled), nil
}

func _getSkippedState(reciever starlark.Value) (starlark.Value, error) {
	return starlark.String(TaskStateSkipped), nil
}
