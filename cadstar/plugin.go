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
	// ID returns unique plugin identifier
	ID() string
	// Create returns a starlark.Value containing plugin's functions and properties.
	// It will be exposed to starlark scripts under the plugin's ID.
	// e.g. ID = random, the plugin will be accessible as random.randint()
	Create(info RunInfo) starlark.Value
	// Register registers Cadence activities if any used by the plugin.
	// Deprecated: Cadence activities have different lifecycle, register them separately, outside the plugin's code.
	Register(registry worker.Registry)
	// LocalStorageKeys returns a list of keys that plugin uses to store data in the local storage.
	// The keys must be unique so prefixing them with plugin name is recommended. e.g. "plugin_name.key"
	// The local storage is shared between all plugins.
	SharedLocalStorageKeys() []string
}
