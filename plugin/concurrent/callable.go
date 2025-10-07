package concurrent

import (
	"fmt"

	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
)

type callable struct {
	Fn   starlark.Callable
	Args starlark.Tuple
}

var (
	_ starlark.HasAttrs = (*callable)(nil)
)

func (r *callable) String() string        { return "concurrent.callable" }
func (r *callable) Type() string          { return "concurrent.callable" }
func (r *callable) Freeze()               {}
func (r *callable) Truth() starlark.Bool  { return true }
func (r *callable) Hash() (uint32, error) { return 0, fmt.Errorf("no-hash") }
func (r *callable) AttrNames() []string   { return star.AttrNames(callableBuiltins, callableProperties) }
func (r *callable) Attr(n string) (starlark.Value, error) {
	return star.Attr(r, n, callableBuiltins, callableProperties)
}

var callableBuiltins = map[string]*starlark.Builtin{}

var callableProperties = map[string]star.PropertyFactory{}

func newCallable(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	fn := args[0].(starlark.Callable)
	args = args[1:]
	return &callable{Fn: fn, Args: args}, nil
}
