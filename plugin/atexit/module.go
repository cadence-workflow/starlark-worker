package atexit

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"github.com/cadence-workflow/starlark-worker/star"
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
	"register":   starlark.NewBuiltin("register", register),
	"unregister": starlark.NewBuiltin("unregister", unregister),
}

var properties = map[string]star.PropertyFactory{}

func register(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	fn := args[0].(starlark.Callable)
	args = args[1:]
	cadstar.GetExitHooks(ctx).Register(fn, args, kwargs)
	return starlark.None, nil
}

func unregister(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	fn := args[0].(starlark.Callable)
	cadstar.GetExitHooks(ctx).Unregister(fn)
	return starlark.None, nil
}
