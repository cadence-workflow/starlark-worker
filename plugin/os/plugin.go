package os

import (
	"github.com/cadence-workflow/starlark-worker/service"
	"github.com/cadence-workflow/starlark-worker/worker"
	"go.starlark.net/starlark"
)

const pluginID = "os"

var Plugin = &plugin{}

type plugin struct{}

var _ service.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(info service.RunInfo) starlark.Value {
	return &Module{environ: info.Environ}
}

func (r *plugin) Register(registry worker.Registry) {}
