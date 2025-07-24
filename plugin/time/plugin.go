package time

import (
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/worker"
	"go.starlark.net/starlark"
)

const pluginID = "time"

var Plugin = &plugin{}

type plugin struct{}

var _ service.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(info service.RunInfo) starlark.Value {
	return NewModule(info)
}

func (r *plugin) Register(registry worker.Registry) {}
