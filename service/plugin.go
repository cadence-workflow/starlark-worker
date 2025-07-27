package service

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/worker"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"go.starlark.net/starlark"
	"strconv"
	"strings"
	"time"
)

// STARLARK_TIME is an environment variable used to alter the time for the Starlark script execution.
// This feature supports backfill scenarios, enabling users to execute their Starlark scripts as if they were triggered
// at a specific time in the past.
//
// The STARLARK_TIME value must be in the following format: 'unix:<value>'.
// Here, "unix" is the scheme, and <value> is a number of seconds since the epoch, commonly known as Unix time.
// Currently, we only support Unix time format. However, we can add support for other formats, like RFC 3339.
//
// See RunInfo.GetEffectiveTime for usage details.
const STARLARK_TIME = "STARLARK_TIME"

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
	SysTime time.Time // SysTime represents the actual wall-clock time when the workflow execution started.
}

// GetEffectiveTime returns the effective time when the workflow execution started.
// By default, it corresponds to SysTime, i.e., actual wall-clock time.
// However, users have the option to alter the time via the STARLARK_TIME environment variable.
// In this case, the function returns altered time as defined by the STARLARK_TIME environment variable.
func (r *RunInfo) GetEffectiveTime() (time.Time, error) {
	v, ok, _ := r.Environ.Get(starlark.String(STARLARK_TIME))
	if !ok {
		return r.SysTime, nil // Return system time if STARLARK_TIME is undefined.
	}
	// Parse STARLARK_TIME. Expected format: scheme:value
	parts := strings.SplitN(v.(starlark.String).GoString(), ":", 2)
	if len(parts) != 2 {
		err := fmt.Errorf("invalid %s env; expected format: 'scheme:value'; actual value: '%s'", STARLARK_TIME, v)
		return time.Time{}, err
	}
	scheme := parts[0]
	rest := parts[1]
	switch scheme {
	case "unix":
		seconds, err := strconv.ParseInt(rest, 10, 64)
		if err != nil {
			return time.Time{}, err
		}
		return time.Unix(seconds, 0), nil
	default:
		err := fmt.Errorf("unsupported scheme: %s", scheme)
		return time.Time{}, err
	}
}
