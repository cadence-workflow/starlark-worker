package os

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
)

type Module struct {
	environ *starlark.Dict
}

var _ starlark.HasAttrs = &Module{}

func (f *Module) String() string                        { return pluginID }
func (f *Module) Type() string                          { return pluginID }
func (f *Module) Freeze()                               {}
func (f *Module) Truth() starlark.Bool                  { return true }
func (f *Module) Hash() (uint32, error)                 { return 0, fmt.Errorf("no-hash") }
func (f *Module) Attr(n string) (starlark.Value, error) { return star.Attr(f, n, builtins, properties) }
func (f *Module) AttrNames() []string                   { return star.AttrNames(builtins, properties) }

var builtins = map[string]*starlark.Builtin{
	"getenv": starlark.NewBuiltin("getenv", _getenv),
}

var properties = map[string]star.PropertyFactory{
	"environ": environ,
}

func environ(receiver starlark.Value) (starlark.Value, error) {
	return receiver.(*Module).environ, nil
}

// _getenv returns the value for the given key from the pipeline run environ,
// or default if the key is absent.
//
// Returns: str or default
func _getenv(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	m := fn.Receiver().(*Module)
	var key string
	defaultVal := starlark.Value(starlark.None)
	if err := starlark.UnpackArgs("getenv", args, kwargs, "key", &key, "default?", &defaultVal); err != nil {
		return starlark.None, err
	}
	v, found, err := m.environ.Get(starlark.String(key))
	if err != nil {
		return starlark.None, err
	}
	if !found {
		return defaultVal, nil
	}
	return v, nil
}


