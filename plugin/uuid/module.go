package uuid

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"github.com/cadence-workflow/starlark-worker/workflow"
	_uuid "github.com/google/uuid"
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
	"uuid4": starlark.NewBuiltin("uuid4", uuid4),
}

var properties = map[string]star.PropertyFactory{}

func uuid4(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := service.GetContext(t)
	logger := workflow.GetLogger(ctx)

	if err := starlark.UnpackArgs("uuid4", args, kwargs); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	_stringUUID := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		return _uuid.New().String()
	})
	var stringUUID starlark.String
	if err := _stringUUID.Get(&stringUUID); err != nil {
		logger.Error("get side effect for uuid4 failed", zap.Error(err))
		return nil, err
	}
	return &UUID{StringUUID: stringUUID}, nil
}
