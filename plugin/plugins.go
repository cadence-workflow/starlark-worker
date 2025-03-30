package plugin

import (
	"github.com/cadence-workflow/starlark-worker/internal/workflow"
	"github.com/cadence-workflow/starlark-worker/plugin/atexit"
	"github.com/cadence-workflow/starlark-worker/plugin/cad"
	"github.com/cadence-workflow/starlark-worker/plugin/concurrent"
	"github.com/cadence-workflow/starlark-worker/plugin/hashlib"
	"github.com/cadence-workflow/starlark-worker/plugin/json"
	"github.com/cadence-workflow/starlark-worker/plugin/os"
	"github.com/cadence-workflow/starlark-worker/plugin/progress"
	"github.com/cadence-workflow/starlark-worker/plugin/random"
	"github.com/cadence-workflow/starlark-worker/plugin/request"
	"github.com/cadence-workflow/starlark-worker/plugin/test"
	"github.com/cadence-workflow/starlark-worker/plugin/time"
	"github.com/cadence-workflow/starlark-worker/plugin/uuid"
)

var Registry = map[string]workflow.IPlugin{
	cad.Plugin.ID():        cad.Plugin,
	request.Plugin.ID():    request.Plugin,
	time.Plugin.ID():       time.Plugin,
	test.Plugin.ID():       test.Plugin,
	os.Plugin.ID():         os.Plugin,
	json.Plugin.ID():       json.Plugin,
	uuid.Plugin.ID():       uuid.Plugin,
	concurrent.Plugin.ID(): concurrent.Plugin,
	atexit.Plugin.ID():     atexit.Plugin,
	progress.Plugin.ID():   progress.Plugin,
	hashlib.Plugin.ID():    hashlib.Plugin,
	random.Plugin.ID():     random.Plugin,
}
