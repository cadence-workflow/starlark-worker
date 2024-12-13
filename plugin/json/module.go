package json

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
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
	"dumps": starlark.NewBuiltin("dumps", dumps),
	"loads": starlark.NewBuiltin("loads", loads),
}

var properties = map[string]star.PropertyFactory{}

func dumps(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	// dumps(obj)
	// Serialize `obj` to a JSON formatted `str`

	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var obj starlark.Value

	if err := starlark.UnpackArgs("dumps", args, kwargs, "obj", &obj); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	encoded, err := star.Encode(obj)
	if err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}
	return starlark.String(encoded), nil
}

func loads(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {

	// Deserialize `s` (a `string` or `bytes` instance containing a JSON document)
	// to a Python object.

	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var s starlark.Value

	if err := starlark.UnpackArgs("loads", args, kwargs, "s", &s); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	var sb []byte
	switch s := s.(type) {
	case starlark.String:
		sb = []byte(s)
	case starlark.Bytes:
		sb = []byte(s)
	default:
		code := "bad-request"
		details := fmt.Sprintf("argument must be a string or bytes; actual: %T: %s", s, s.String())
		logger.Error(code, zap.String("details", details))
		return nil, cadence.NewCustomError(code, details)
	}

	var res starlark.Value
	if err := star.Decode(sb, &res); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}
	return res, nil
}
