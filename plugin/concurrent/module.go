package concurrent

import (
	"fmt"
	pworkflow "github.com/cadence-workflow/starlark-worker/plugin/workflow"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.starlark.net/starlark"
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
	"run": starlark.NewBuiltin("run", run),
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
