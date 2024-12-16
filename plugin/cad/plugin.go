package cad

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/worker"
)

const pluginID = "cad"

var Plugin = &plugin{}

type plugin struct{}

var _ cadstar.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(info cadstar.RunInfo) starlark.Value {
	return &Module{info: info.Info}
}

func (r *plugin) Register(registry worker.Registry) {}
