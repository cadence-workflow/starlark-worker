package cad

import (
	"github.com/cadence-workflow/starlark-worker/internal/worker"
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"go.starlark.net/starlark"
)

const pluginID = "cad"

var Plugin = &plugin{}

type plugin struct{}

var _ workflow.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(info workflow.RunInfo) starlark.Value {
	return &Module{info: info.Info}
}

func (r *plugin) Register(registry worker.Registry) {}
