package random

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/worker"
)

const (
	pluginID = "random"
	// threadLocalSeededRandInstanceKey is the key to store the *rand.Rand instance in the thread local storage
	// which gets created when random.seed(x) is called.
	threadLocalSeededRandInstanceKey = pluginID + ".seeded_instance"
)

var Plugin = &plugin{}

type plugin struct{}

var _ cadstar.IPlugin = (*plugin)(nil)

func (r *plugin) Create(info cadstar.RunInfo) starlark.StringDict {
	return starlark.StringDict{pluginID: &Module{}}
}

func (r *plugin) Register(registry worker.Registry) {}

func (r *plugin) SharedLocalStorageKeys() []string {
	return []string{threadLocalSeededRandInstanceKey}
}
