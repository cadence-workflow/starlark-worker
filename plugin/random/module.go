package random

import (
	"fmt"
	"github.com/cadence-workflow/starlark-worker/ext"
	"math/rand"

	"github.com/cadence-workflow/starlark-worker/cadstar"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type Module struct {
	attributes map[string]starlark.Value
	rand       *rand.Rand
}

func NewModule() *Module {
	m := &Module{}
	m.attributes = map[string]starlark.Value{
		"seed":    starlark.NewBuiltin("seed", m.seedFn).BindReceiver(m),
		"randint": starlark.NewBuiltin("randint", m.randIntFn).BindReceiver(m),
		"random":  starlark.NewBuiltin("random", m.randFn).BindReceiver(m),
	}
	return m
}

var _ starlark.HasAttrs = &Module{}

func (f *Module) String() string                        { return pluginID }
func (f *Module) Type() string                          { return pluginID }
func (f *Module) Freeze()                               {}
func (f *Module) Truth() starlark.Bool                  { return true }
func (f *Module) Hash() (uint32, error)                 { return 0, fmt.Errorf("no-hash") }
func (f *Module) Attr(n string) (starlark.Value, error) { return f.attributes[n], nil }
func (f *Module) AttrNames() []string                   { return ext.SortedKeys(f.attributes) }

// seedFn is used to seed the random number generator
// Arguments:
//   - seed: the seed value
//
// Return: None
func (f *Module) seedFn(t *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var seed int64
	if err := starlark.UnpackArgs("int", args, kwargs, "seed", &seed); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	// create a new random source with the seed
	f.rand = rand.New(rand.NewSource(int64(seed)))

	return starlark.None, nil
}

// randIntFn generates a random integer between the specified range
// Arguments:
//   - min: the minimum value of the random integer
//   - max: the maximum value of the random integer
//
// Return: The generated random integer
func (f *Module) randIntFn(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var min, max int
	if err := starlark.UnpackArgs("int", args, kwargs, "min", &min, "max", &max); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	var v int
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		if f.rand != nil { // seed was called before and a random source was stored in the thread local. Use it
			return f.rand.Intn(max-min+1) + min
		}

		// seed was not called before. Use the default random source
		return rand.Intn(max-min+1) + min
	}).Get(&v)
	if err != nil {
		logger.Error("get side effect for randIntFn failed", zap.Error(err))
		return nil, err
	}

	return starlark.MakeInt(v), nil
}

// randFn generates a random floating point number between 0 and 1
// Return: The generated random number
func (f *Module) randFn(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var v float64
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		if f.rand != nil { // seed was called before and a random source was stored in the thread local. Use it
			return f.rand.Float64()
		}

		// seed was not called before. Use the default random source
		return rand.Float64()
	}).Get(&v)
	if err != nil {
		logger.Error("get side effect for randIntFn failed", zap.Error(err))
		return nil, err
	}

	return starlark.Float(v), nil
}
