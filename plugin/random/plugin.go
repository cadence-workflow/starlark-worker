package random

import (
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/worker"
	"go.starlark.net/starlark"
)

const (
	pluginID = "random"
)

var Plugin = &plugin{}

type plugin struct{}

var _ service.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(_ service.RunInfo) starlark.Value {
	return NewModule()
}

func (r *plugin) Register(registry worker.Registry) {}
