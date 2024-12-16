package cadstar

import (
	"container/list"
	"errors"
	"fmt"
	extcadence "github.com/cadence-workflow/starlark-worker/cad"
	"github.com/cadence-workflow/starlark-worker/ext"

	"github.com/cadence-workflow/starlark-worker/star"
	jsoniter "github.com/json-iterator/go"
	"go.starlark.net/starlark"
	"go.uber.org/cadence"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
	"go.uber.org/yarpc/yarpcerrors"
	"go.uber.org/zap"
)

type contextKey int

const contextKeyGlobals contextKey = iota

var builtins = starlark.StringDict{
	star.CallableObjectType: star.CallableObjectConstructor,
	star.DataclassType:      star.DataclassConstructor,
}

type Meta struct {
	MainFile     string `json:"main_file,omitempty"`
	MainFunction string `json:"main_function,omitempty"`
}

type _Globals struct {
	exitHooks  *ExitHooks
	isCanceled bool
	logs       *list.List
	environ    *starlark.Dict
	progress   *list.List
	plugins    map[string]IPlugin
}

func (r *_Globals) getEnviron(key starlark.String) (starlark.String, bool) {
	v, found, err := r.environ.Get(key)
	if err != nil {
		panic(err)
	}
	if !found {
		return "", false
	}
	return v.(starlark.String), true
}

type Service struct {
	Plugins        map[string]IPlugin
	ClientTaskList string
}

// TODO: [feature] Cadence workflow with starlark REPL (event listener loop?) starlark.ExecREPLChunk()

func (r *Service) Run(
	ctx workflow.Context,
	tar []byte,
	path string,
	function string,
	args starlark.Tuple,
	kwargs []starlark.Tuple,
	environ *starlark.Dict,
) (
	res starlark.Value,
	err error,
) {

	logger := workflow.GetLogger(ctx)

	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("workflow-panic", zap.Any("panic", rec))
			err = cadence.NewCustomError(
				yarpcerrors.CodeInternal.String(),
				fmt.Sprintf("panic: %v", rec),
			)
		}
	}()

	logger.Info(
		"workflow-start",
		zap.String("path", path),
		zap.String("function", function),
		zap.Int("tar_len", len(tar)),
	)

	if environ == nil {
		environ = &starlark.Dict{}
	}

	ao := extcadence.DefaultActivityOptions
	ao.TaskList = r.ClientTaskList
	ctx = workflow.WithActivityOptions(ctx, ao)

	cwo := extcadence.DefaultChildWorkflowOptions
	cwo.TaskList = r.ClientTaskList
	ctx = workflow.WithChildOptions(ctx, cwo)

	globals := &_Globals{
		exitHooks:  &ExitHooks{},
		isCanceled: false,
		logs:       list.New(),
		environ:    environ,
		progress:   list.New(),
		plugins:    r.Plugins,
	}
	ctx = workflow.WithValue(ctx, contextKeyGlobals, globals)

	var fs star.FS
	if fs, err = star.NewTarFS(tar); err != nil {
		logger.Error("workflow-error", ext.ZapError(err)...)
		return nil, cadence.NewCustomError(
			yarpcerrors.CodeInvalidArgument.String(),
			err.Error(),
		)
	}

	meta := Meta{}
	if b, err := fs.Read("/meta.json"); err != nil {
		if !errors.Is(err, star.ErrNotExist) {
			return nil, err
		}
	} else {
		if err := jsoniter.Unmarshal(b, &meta); err != nil {
			return nil, err
		}
	}
	logger.Info("workflow-meta", zap.Any("meta", meta))

	if path == "" {
		path = meta.MainFile
	}
	if function == "" {
		function = meta.MainFunction
	}

	runInfo := RunInfo{
		Info:    workflow.GetInfo(ctx),
		Environ: environ,
	}

	plugins := starlark.StringDict{}
	for pID, p := range r.Plugins {
		plugins[pID] = p.Create(runInfo)
	}

	if err := workflow.SetQueryHandler(ctx, "logs", func() (any, error) {
		logs := make([]any, globals.logs.Len())
		var i int
		for e := globals.logs.Front(); e != nil; e = e.Next() {
			logs[i] = e.Value
			i++
		}
		return logs, nil
	}); err != nil {
		logger.Error("workflow-error", ext.ZapError(err)...)
		return nil, err
	}

	if err := workflow.SetQueryHandler(ctx, "task_progress", func() (any, error) {
		progress := make([]any, globals.progress.Len())
		var i int
		for e := globals.progress.Front(); e != nil; e = e.Next() {
			progress[i] = e.Value
			i++
		}
		return progress, nil
	}); err != nil {
		logger.Error("workflow-error", ext.ZapError(err)...)
		return nil, err
	}

	t := CreateThread(ctx)
	t.Load = star.ThreadLoad(fs, builtins, map[string]starlark.StringDict{"plugin": plugins})

	// Run main user code
	if res, err = star.Call(t, path, function, args, kwargs); err != nil {
		logger.Error("workflow-error", ext.ZapError(err)...)

		var canceledError *cadence.CanceledError
		if errors.As(err, &canceledError) {
			globals.isCanceled = true
			ctx, _ = workflow.NewDisconnectedContext(ctx)
		}
	}

	// Run exit hooks
	if _err := globals.exitHooks.Run(t); _err != nil {
		logger.Error("exit-hook-error", ext.ZapError(_err)...)
		err = errors.Join(err, _err)
	}

	err = processError(ctx, err)

	if err != nil {
		exec := workflow.GetInfo(ctx).WorkflowExecution
		tags := map[string]string{
			"w_id":   exec.ID,
			"run_id": exec.RunID,
			"error":  err.Error(),
		}
		workflow.GetMetricsScope(ctx).Tagged(tags).Gauge("workflow.error").Update(1)
	}
	logger.Info("workflow-end")
	return res, err
}

func processError(ctx workflow.Context, err error) error {
	if err == nil {
		return nil
	}
	logger := workflow.GetLogger(ctx)

	details := map[string]any{"error": err.Error()}
	var evalErr *starlark.EvalError
	if errors.As(err, &evalErr) {
		logger.Error("starlark-backtrace", zap.String("backtrace", evalErr.Backtrace()))
		details["backtrace"] = evalErr.Backtrace()
	}
	var cadenceErr *cadence.CustomError
	var reason = yarpcerrors.CodeUnknown.String()
	if errors.As(err, &cadenceErr) {
		reason = cadenceErr.Reason()
		if cadenceErr.HasDetails() {
			var d any
			if err := cadenceErr.Details(&d); err != nil {
				logger.Error("workflow-error", ext.ZapError(err)...)
				d = fmt.Sprintf("internal: error details extraction failure: %s", err.Error())
			}
			details["details"] = d
		}
	}
	return cadence.NewCustomError(reason, details)
}

func (r *Service) Register(registry worker.Registry) {
	registry.RegisterWorkflow(r.Run)
	for _, plugin := range r.Plugins {
		plugin.Register(registry)
	}
}

func GetExitHooks(ctx workflow.Context) *ExitHooks {
	return getGlobals(ctx).exitHooks
}

func getGlobals(ctx workflow.Context) *_Globals {
	return ctx.Value(contextKeyGlobals).(*_Globals)
}

func GetProgress(ctx workflow.Context) *list.List {
	return getGlobals(ctx).progress
}
