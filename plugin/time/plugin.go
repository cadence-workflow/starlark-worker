package time

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
)

var Plugin = cadstar.PluginFactory(func(info cadstar.RunInfo) starlark.StringDict {
	return starlark.StringDict{"time": &Module{}}
})
