package time

import (
	"fmt"
	"strings"
	"time"

	"github.com/cadence-workflow/starlark-worker/cadstar"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
	"go.uber.org/cadence"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type Module struct{}

var _ starlark.HasAttrs = &Module{}

func (f *Module) String() string                        { return "time" }
func (f *Module) Type() string                          { return "time" }
func (f *Module) Freeze()                               {}
func (f *Module) Truth() starlark.Bool                  { return true }
func (f *Module) Hash() (uint32, error)                 { return 0, fmt.Errorf("no-hash") }
func (f *Module) Attr(n string) (starlark.Value, error) { return star.Attr(f, n, builtins, properties) }
func (f *Module) AttrNames() []string                   { return star.AttrNames(builtins, properties) }

func (f *Module) Sleep(t *starlark.Thread, seconds starlark.Value) error {
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

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
		return cadence.NewCustomError(code, details)
	}

	return workflow.Sleep(ctx, time.Duration(float64(time.Second)*sf))
}

var builtins = map[string]*starlark.Builtin{
	"sleep":              starlark.NewBuiltin("sleep", _sleep),
	"time_ns":            starlark.NewBuiltin("time_ns", _time_ns),
	"time":               starlark.NewBuiltin("time", _time),
	"utc_format_seconds": starlark.NewBuiltin("utc_format_seconds", _utc_format_seconds),
}

var properties = map[string]star.PropertyFactory{}

// _sleep suspends execution of the calling thread for the given number of seconds.
// The argument may be a floating point number to indicate a more precise sleep time.
// Arguments:
//   - seconds: the number of seconds to sleep.
//
// Returns: None
func _sleep(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	r := fn.Receiver().(*Module)
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var seconds starlark.Value
	if err := starlark.UnpackArgs("sleep", args, kwargs, "seconds", &seconds); err != nil {
		logger.Error("builtin-error", ext.ZapError(err)...)
		return nil, err
	}
	return starlark.None, r.Sleep(t, seconds)
}

// _time_ns is similar to _time but returns time as an integer number of nanoseconds since the epoch.
// Returns: int
func _time_ns(t *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	ns := workflow.Now(ctx).UnixNano()
	return starlark.MakeInt64(ns), nil
}

// _time returns the current unix time in seconds as floating point number.
// Use _time_ns to avoid the precision loss caused by the float type.
// Returns: float
func _time(t *starlark.Thread, _ *starlark.Builtin, _ starlark.Tuple, _ []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	ns := workflow.Now(ctx).UnixNano()
	sec := float64(ns) / 1e9
	return starlark.Float(sec), nil
}

// _utc_format_seconds converts the given unix time in seconds to a string as specified by the format argument.
// The formatted result string represents the UTC time. Arguments:
//   - format: the format string containing the date and time directives such as %Y, %m, %d, %H, %M, %S.
//   - seconds: the unix time in seconds.
//
// Returns: str
func _utc_format_seconds(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kw []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
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
		return nil, cadence.NewCustomError("400", fmt.Sprintf("unsupported date format: %s", format))
	}

	res := time.Unix(int64(seconds), 0).UTC().Format(format)
	return starlark.String(res), nil
}
