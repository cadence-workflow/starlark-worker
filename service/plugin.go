package service

import (
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.starlark.net/starlark"
)

// IPlugin plugin factory interface
// Plugin instances are created on startup and are used to create starlark.Value instances per workflow execution
type IPlugin interface {
	// ID returns unique plugin identifier
	ID() string
	// Create returns a starlark.Value containing plugin's functions and properties.
	// It will be exposed to starlark scripts under the plugin's ID.
	// e.g. ID = random, the plugin will be accessible as random.randint()
	Create(info RunInfo) starlark.Value
	// Register registers Cadence activities if any used by the plugin.
	Register(registry worker.Registry)
}

// RunInfo contextual info about the current run
type RunInfo struct {
	Info    workflow.IInfo
	Environ *starlark.Dict
}
