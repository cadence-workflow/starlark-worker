package time

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/workflow"
	"strings"
	"time"

	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/service"
	"go.starlark.net/starlark"
	"go.uber.org/zap"
)

type Module struct {
	attributes map[string]starlark.Value
	// delta is the duration applied to the current system time to resolve and return the effective time to the caller.
	//
	// See getEffectiveTime.
	delta time.Duration
}

func NewModule(info service.RunInfo) starlark.Value {
	m := &Module{}
	m.attributes = map[string]starlark.Value{
		"sleep":              starlark.NewBuiltin("sleep", m._sleep).BindReceiver(m),
		"time_ns":            starlark.NewBuiltin("time_ns", m._time_ns).BindReceiver(m),
		"time":               starlark.NewBuiltin("time", m._time).BindReceiver(m),
		"utc_format_seconds": starlark.NewBuiltin("utc_format_seconds", m._utc_format_seconds).BindReceiver(m),
	}
	effectiveTime := ext.Must(info.GetEffectiveTime())
	m.delta = effectiveTime.Sub(info.SysTime)
	return m
}

var _ starlark.HasAttrs = &Module{}

func (m *Module) String() string                        { return pluginID }
func (m *Module) Type() string                          { return pluginID }
func (m *Module) Freeze()                               {}
func (m *Module) Truth() starlark.Bool                  { return true }
func (m *Module) Hash() (uint32, error)                 { return 0, fmt.Errorf("no-hash") }
func (m *Module) Attr(n string) (starlark.Value, error) { return m.attributes[n], nil }
func (m *Module) AttrNames() []string                   { return ext.SortedKeys(m.attributes) }

// _sleep suspends execution of the calling thread for the given number of seconds.
// The argument may be a floating point number to indicate a more precise sleep time.
// Arguments:
//   - seconds: the number of seconds to sleep.
//
// Returns: None
func (m *Module) _sleep(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := service.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var seconds starlark.Value
	if err := starlark.UnpackArgs("sleep", args, kwargs, "seconds", &seconds); err != nil {
		logger.Error("builtin-error", ext.ZapError(err)...)
		return nil, err
	}

	var sf float64
	switch arg0 := seconds.(type) {
	case starlark.Int:
		sf = float64(arg0.Float())
	case starlark.Float:
		sf = float64(arg0)
	default:
		code := "bad-request"
		details := fmt.Sprintf("bad argument type: %T: %v", seconds, seconds)
		logger.Error(code, zap.String("details", details))
		return starlark.None, workflow.NewCustomError(ctx, code, details)
	}

	return starlark.None, workflow.Sleep(ctx, time.Duration(float64(time.Second)*sf))
}

// _time_ns is similar to _time but returns time as an integer number of nanoseconds since the epoch.
// Returns: int
func (m *Module) _time_ns(t *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	ns := m.getEffectiveTime(t).UnixNano()
	return starlark.MakeInt64(ns), nil
}

// _time returns the current unix time in seconds as floating point number.
// Use _time_ns to avoid the precision loss caused by the float type.
// Returns: float
func (m *Module) _time(t *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	ns := m.getEffectiveTime(t).UnixNano()
	sec := float64(ns) / 1e9
	return starlark.Float(sec), nil
}

// _utc_format_seconds converts the given unix time in seconds to a string as specified by the format argument.
// The formatted result string represents the UTC time. Arguments:
//   - format: the format string containing the date and time directives such as %Y, %m, %d, %H, %M, %S.
//   - seconds: the unix time in seconds.
//
// Returns: str
func (m *Module) _utc_format_seconds(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	ctx := service.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var format string
	var seconds float64
	if err := starlark.UnpackArgs("format_time", args, kw,
		"format", &format,
		"seconds", &seconds,
	); err != nil {
		logger.Error("builtin-error", ext.ZapError(err)...)
		return nil, err
	}

	replacer := strings.NewReplacer("%Y", "2006", "%m", "01", "%d", "02", "%H", "15", "%M", "04", "%S", "05", "%y", "06")
	format = replacer.Replace(format)
	if strings.Contains(format, "%") {
		return nil, workflow.NewCustomError(ctx, "400", fmt.Sprintf("unsupported date format: %s", format))
	}

	res := time.Unix(int64(seconds), 0).UTC().Format(format)
	return starlark.String(res), nil
}

// getEffectiveTime returns the current time, taking into account the delta duration, which is non-zero,
// if the time is altered via the service.STARLARK_TIME environment variable.
func (m *Module) getEffectiveTime(t *starlark.Thread) time.Time {
	ctx := service.GetContext(t)
	return workflow.Now(ctx).Add(m.delta)
}
