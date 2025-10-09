package concurrent

import (
	"fmt"

	pworkflow "github.com/cadence-workflow/starlark-worker/plugin/workflow"
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

var builtins = map[string]*starlark.Builtin{
	"run":          starlark.NewBuiltin("run", run),
	"batch_run":    starlark.NewBuiltin("batch_run", batchRun),
	"new_callable": starlark.NewBuiltin("new_callable", newCallable),
}

var properties = map[string]star.PropertyFactory{}

func run(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var ctx = service.GetContext(t)
	future, settable := workflow.NewFuture(ctx)
	fn := args[0]
	args = args[1:]
	workflow.Go(ctx, func(ctx workflow.Context) {
		subT := service.CreateThread(ctx)
		settable.Set(starlark.Call(subT, fn, args, kwargs))
	})
	return &pworkflow.Future{Future: future}, nil
}

func batchRun(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := service.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var callablesList *starlark.List
	var maxConcurrency int
	if err := starlark.UnpackArgs("batchRun", args, kwargs, "callables", &callablesList, "max_concurrency?", &maxConcurrency); err != nil {
		logger.Error("error unpacking args:", zap.Error(err))
		return nil, err
	}

	// Convert starlark.List to slice of callable objects
	callables := make([]*callable, callablesList.Len())
	for i := 0; i < callablesList.Len(); i++ {
		item := callablesList.Index(i)
		if callableObj, ok := item.(*callable); ok {
			callables[i] = callableObj
		} else {
			err := fmt.Errorf("list item %d is not a callable, got %T", i, item)
			logger.Error("error unpacking args:", zap.Error(err))
			return nil, err
		}
	}

	// if maxConcurrency is not provided, start all jobs in parallel
	if maxConcurrency == 0 {
		maxConcurrency = len(callables)
	}

	// construct factories for each callable
	factories := make([]func(ctx workflow.Context) workflow.Future, len(callables))
	for i, callableObj := range callables {
		callableObj := callableObj
		factories[i] = func(ctx workflow.Context) workflow.Future {
			fn := callableObj.Fn
			args := callableObj.Args
			future, settable := workflow.NewFuture(ctx)
			workflow.Go(ctx, func(ctx workflow.Context) {
				subT := service.CreateThread(ctx)
				settable.Set(starlark.Call(subT, fn, args, nil))
			})
			return future
		}
	}

	batchFuture, err := workflow.NewBatchFuture(ctx, maxConcurrency, factories)
	if err != nil {
		logger.Error("error creating batch future:", zap.Error(err))
		return nil, err
	}

	return &pworkflow.BatchFuture{BatchFuture: batchFuture}, nil
}
