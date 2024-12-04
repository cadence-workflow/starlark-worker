package cadstar

import (
	"go.starlark.net/starlark"
	"go.uber.org/cadence/worker"
	"go.uber.org/cadence/workflow"
)

// RunInfo contextual info about the current run
type RunInfo struct {
	Info    *workflow.Info
	Environ *starlark.Dict
}

// IPlugin plugin factory interface
type IPlugin interface {
	// Create instantiates starlark.StringDict that exposes plugin's functions and properties
	Create(info RunInfo) starlark.StringDict
	// Register registers Cadence activities if any used by the plugin.
	// Deprecated: Cadence activities have different lifecycle, register them separately, outside the plugin's code.
	Register(registry worker.Registry)
}

// PluginFactory is a functional IPlugin implementation
type PluginFactory func(info RunInfo) starlark.StringDict

var _ IPlugin = (PluginFactory)(nil)

func (r PluginFactory) Create(info RunInfo) starlark.StringDict {
	return r(info)
}

// Register does nothing.
// Deprecated: Cadence activities have different lifecycle, register them separately, outside the plugin's code.
func (r PluginFactory) Register(_ worker.Registry) {}
