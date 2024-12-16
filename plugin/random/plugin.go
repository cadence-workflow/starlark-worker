package random

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/worker"
)

const (
	pluginID = "random"
)

var Plugin = &plugin{}

type plugin struct{}

var _ cadstar.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(_ cadstar.RunInfo) starlark.Value {
	return NewModule()
}

func (r *plugin) Register(registry worker.Registry) {}
