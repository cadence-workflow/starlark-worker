package plugin

import (
	"github.com/cadence-workflow/starlark-worker/cadstar"
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

var Registry = []cadstar.IPlugin{
	cad.Plugin,
	request.Plugin,
	time.Plugin,
	test.Plugin,
	os.Plugin,
	json.Plugin,
	uuid.Plugin,
	concurrent.Plugin,
	atexit.Plugin,
	progress.Plugin,
	hashlib.Plugin,
	random.Plugin,
}
