package test

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/worker"
)

const pluginID = "test"

var Plugin = &plugin{}

type plugin struct{}

var _ cadstar.IPlugin = (*plugin)(nil)

func (r *plugin) ID() string {
	return pluginID
}

func (r *plugin) Create(info cadstar.RunInfo) starlark.StringDict {
	return starlark.StringDict{pluginID: &Module{}}
}

func (r *plugin) Register(registry worker.Registry) {}

func (r *plugin) SharedLocalStorageKeys() []string { return nil }
