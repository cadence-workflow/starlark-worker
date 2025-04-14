package request

import (
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/worker"
	"go.starlark.net/starlark"
	"net/http"
)

const pluginID = "request"

var Plugin = &plugin{}

type plugin struct{}

var _ service.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(_ service.RunInfo) starlark.Value {
	return &Module{}
}

func (r *plugin) Register(registry worker.Registry) {
	registry.RegisterActivity(&activities{
		client: http.DefaultClient,
	})
}
