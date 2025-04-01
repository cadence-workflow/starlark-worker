package request

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
	"net/http"
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
	"do": starlark.NewBuiltin("do", _do),
}

var properties = map[string]star.PropertyFactory{}

func _do(t *starlark.Thread, b *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := service.GetContext(t)
	w := service.GetWorkflow(t)
	logger := w.GetLogger(ctx)

	var method starlark.String
	var url starlark.String
	var body starlark.Value = starlark.None
	var headers starlark.Value = starlark.None

	if err := starlark.UnpackArgs("do", args, kwargs, "method", &method, "url", &url, "body?", &body, "headers?", &headers); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}
	var err error
	var res *http.Response
	future := w.ExecuteActivity(ctx, Activities.Do, method, url, headers, body)
	if res, err = getResponse(ctx, future); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}
	return &Response{Response: res}, nil
}

func getResponse(ctx workflow.Context, future workflow.Future) (*http.Response, error) {
	var b []byte
	if err := future.Get(ctx, &b); err != nil {
		return nil, err
	}
	return http.ReadResponse(bufio.NewReader(bytes.NewBuffer(b)), nil)
}
