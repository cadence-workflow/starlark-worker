package service

import (
	"container/list"
	"errors"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/cadence"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/temporal"
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	jsoniter "github.com/json-iterator/go"
	"github.com/uber-go/tally"
	"go.starlark.net/starlark"
	"go.temporal.io/sdk/client"
	"go.uber.org/yarpc/yarpcerrors"
	"go.uber.org/zap"
)

type BackendType string

const (
	ServiceWorkflowFunc             = "starlark-worklow"
	CadenceBackend      BackendType = "cadence"
	TemporalBackend     BackendType = "temporal"
)

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
	Plugins              map[string]IPlugin
	ActivityOptions      workflow.ActivityOptions
	ChildWorkflowOptions workflow.ChildWorkflowOptions

	workflow workflow.Workflow
}

// NewService creates a service with default options
// Deprecated: Use NewServiceBuilder instead for better extensibility and configuration
func NewService(plugins map[string]IPlugin, clientTaskList string, backendType BackendType) (*Service, error) {
	return NewServiceBuilder(backendType).
		SetPlugins(plugins).
		SetClientTaskList(clientTaskList).
		Build()
}

// ServiceBuilder provides a builder pattern for creating Service instances
type ServiceBuilder struct {
	backendType          BackendType
	plugins              map[string]IPlugin
	activityOptions      workflow.ActivityOptions
	childWorkflowOptions workflow.ChildWorkflowOptions
}

// NewServiceBuilder creates a new ServiceBuilder with the specified backend type
func NewServiceBuilder(backendType BackendType) *ServiceBuilder {
	return &ServiceBuilder{
		backendType:          backendType,
		plugins:              make(map[string]IPlugin),
		activityOptions:      DefaultActivityOptions,
		childWorkflowOptions: DefaultChildWorkflowOptions,
	}
}

// SetPlugins sets the plugin registry
func (b *ServiceBuilder) SetPlugins(plugins map[string]IPlugin) *ServiceBuilder {
	b.plugins = plugins
	return b
}

// SetClientTaskList sets the TaskList for both ActivityOptions and ChildWorkflowOptions
func (b *ServiceBuilder) SetClientTaskList(taskList string) *ServiceBuilder {
	// Override TaskList in both ActivityOptions and ChildWorkflowOptions
	b.activityOptions.TaskList = taskList
	b.childWorkflowOptions.TaskList = taskList
	return b
}

// SetActivityOptions sets the activity options configuration
func (b *ServiceBuilder) SetActivityOptions(options workflow.ActivityOptions) *ServiceBuilder {
	b.activityOptions = options
	return b
}

// SetChildWorkflowOptions sets the child workflow options configuration
func (b *ServiceBuilder) SetChildWorkflowOptions(options workflow.ChildWorkflowOptions) *ServiceBuilder {
	b.childWorkflowOptions = options
	return b
}

// Build creates the Service instance with the configured options
func (b *ServiceBuilder) Build() (*Service, error) {
	var be workflow.Workflow
	switch b.backendType {
	case CadenceBackend:
		be = cadence.NewWorkflow()
	case TemporalBackend:
		be = temporal.NewWorkflow()
	default:
		return nil, fmt.Errorf("unsupported backend: %s", b.backendType)
	}

	// Use configured ActivityOptions and ChildWorkflowOptions
	ao := b.activityOptions
	cwo := b.childWorkflowOptions

	return &Service{
		Plugins:              b.plugins,
		ActivityOptions:      ao,
		ChildWorkflowOptions: cwo,
		workflow:             be,
	}, nil
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
	if r.workflow == nil {
		return nil, fmt.Errorf("backend not initialized")
	}

	ctx = workflow.WithBackend(ctx, r.workflow)

	logger := workflow.GetLogger(ctx)

	defer func() {
		if rec := recover(); rec != nil {
			logger.Error("workflow-panic", zap.Any("panic", rec))
			err = workflow.NewCustomError(
				ctx,
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

	ao := r.ActivityOptions
	ctx = workflow.WithActivityOptions(ctx, ao)

	cwo := r.ChildWorkflowOptions
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
		return nil, workflow.NewCustomError(
			ctx,
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
		SysTime: workflow.Now(ctx),
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

		var canceledErr workflow.CanceledError
		if errors.As(err, &canceledErr) {
			globals.isCanceled = true
			ctx, _ = workflow.NewDisconnectedContext(ctx)
		}
	}

	// Run exit hooks
	if _err := globals.exitHooks.Run(t); _err != nil {
		logger.Error("exit-hook-error", ext.ZapError(_err)...)
		err = errors.Join(err, _err)
	}

	err = r.processError(ctx, err)

	if err != nil {
		exec := workflow.GetInfo(ctx)
		tags := map[string]string{
			"w_id":   exec.ExecutionID(),
			"run_id": exec.RunID(),
			"error":  err.Error(),
		}
		scope := workflow.GetMetricsScope(ctx)
		switch scope.(type) {
		case tally.Scope:
			scope.(tally.Scope).Tagged(tags).Gauge("workflow.error").Update(1)
		case client.MetricsHandler:
			scope.(client.MetricsHandler).WithTags(tags).Gauge("workflow.error").Update(1)
		default:
			logger.Warn("workflow-warning", zap.String("error", "unknown-metrics-scope-type"))
		}
	}
	logger.Info("workflow-end")
	return res, err
}

func (r *Service) processError(ctx workflow.Context, err error) error {
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
	var workflowErr workflow.CustomError
	var reason = yarpcerrors.CodeUnknown.String()
	if errors.As(err, &workflowErr) {
		reason = workflowErr.Reason()
		if workflowErr.HasDetails() {
			var d any
			if err := workflowErr.Details(&d); err != nil {
				logger.Error("workflow-error", ext.ZapError(err)...)
				d = fmt.Sprintf("internal: error details extraction failure: %s", err.Error())
			}
			details["details"] = d
		}
	}
	return workflow.NewCustomError(ctx, reason, details)
}

func (r *Service) Register(registry worker.Registry) {
	registry.RegisterWorkflow(r.Run, ServiceWorkflowFunc)
	for _, plugin := range r.Plugins {
		plugin.Register(registry)
	}
}

func (r *Service) RegisterWithOptions(registry worker.Registry, options worker.RegisterWorkflowOptions) {
	registry.RegisterWorkflowWithOptions(r.Run, options)
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
