package request

import (
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.starlark.net/starlark"
	"net/http"
)

const pluginID = "request"

var Plugin = &plugin{}

type plugin struct{}

var _ workflow.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(_ workflow.RunInfo) starlark.Value {
	return &Module{}
}

func (r *plugin) Register(registry worker.Registry) {
	registry.RegisterActivity(&activities{
		client: http.DefaultClient,
	})
}
