package workflow

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.starlark.net/starlark"
)

type Future struct {
	Future workflow.Future
}

var (
	_ starlark.HasAttrs = (*Future)(nil)
)

func (r *Future) String() string        { return "workflow.future" }
func (r *Future) Type() string          { return "workflow.future" }
func (r *Future) Freeze()               {}
func (r *Future) Truth() starlark.Bool  { return true }
func (r *Future) Hash() (uint32, error) { return 0, fmt.Errorf("no-hash") }
func (r *Future) AttrNames() []string   { return star.AttrNames(futureBuiltins, futureProperties) }
func (r *Future) Attr(n string) (starlark.Value, error) {
	return star.Attr(r, n, futureBuiltins, futureProperties)
}

func (r *Future) Result(t *starlark.Thread) (starlark.Value, error) {
	ctx := service.GetContext(t)
	var res starlark.Value
	if err := r.Future.Get(ctx, &res); err != nil {
		return nil, err
	}
	return res, nil
}

var futureBuiltins = map[string]*starlark.Builtin{
	"result": starlark.NewBuiltin("result", futureResult),
}

var futureProperties = map[string]star.PropertyFactory{}

func futureResult(t *starlark.Thread, fn *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	r := fn.Receiver().(*Future)
	return r.Result(t)
}
