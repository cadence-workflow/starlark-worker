package uuid

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
)

var Plugin = cadstar.PluginFactory(func(info cadstar.RunInfo) starlark.StringDict {
	return starlark.StringDict{"uuid": &Module{}}
})
