package cad

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
)

type Future struct {
	Future workflow.Future
}

var (
	_ starlark.HasAttrs = (*Future)(nil)
)

func (r *Future) String() string        { return "cad.future" }
func (r *Future) Type() string          { return "cad.future" }
func (r *Future) Freeze()               {}
func (r *Future) Truth() starlark.Bool  { return true }
func (r *Future) Hash() (uint32, error) { return 0, fmt.Errorf("no-hash") }
func (r *Future) AttrNames() []string   { return star.AttrNames(futureBuiltins, futureProperties) }
func (r *Future) Attr(n string) (starlark.Value, error) {
	return star.Attr(r, n, futureBuiltins, futureProperties)
}

func (r *Future) Result(t *starlark.Thread) (starlark.Value, error) {
	ctx := service.GetContext(t)

	// Debug: Log thread and context information
	fmt.Printf("[DEBUG] Thread: %s\n", t.Name)
	fmt.Printf("[DEBUG] Future.IsReady: %v\n", r.Future.IsReady())

	var res starlark.Value
	if err := r.Future.Get(ctx, &res); err != nil {
		fmt.Printf("[DEBUG] Future.Get() error: %v\n", err)
		return nil, err
	}

	fmt.Printf("[DEBUG] Future result retrieved: %v\n", res)
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
