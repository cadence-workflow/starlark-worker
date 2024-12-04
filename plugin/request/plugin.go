package request

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/worker"
	"net/http"
)

var Plugin = &plugin{}

type plugin struct{}

var _ cadstar.IPlugin = (*plugin)(nil)

func (r *plugin) Create(_ cadstar.RunInfo) starlark.StringDict {
	return starlark.StringDict{"request": &Module{}}
}

func (r *plugin) Register(registry worker.Registry) {
	registry.RegisterActivity(&activities{
		client: http.DefaultClient,
	})
}
