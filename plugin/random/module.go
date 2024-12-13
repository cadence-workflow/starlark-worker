package random

import (
	"fmt"
	"math/rand"

	"github.com/cadence-workflow/starlark-worker/cadstar"
	"github.com/cadence-workflow/starlark-worker/star"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/workflow"
	"go.uber.org/zap"
)

type Module struct{}

var _ starlark.HasAttrs = &Module{}

func (f *Module) String() string                        { return pluginID }
func (f *Module) Type() string                          { return pluginID }
func (f *Module) Freeze()                               {}
func (f *Module) Truth() starlark.Bool                  { return true }
func (f *Module) Hash() (uint32, error)                 { return 0, fmt.Errorf("no-hash") }
func (f *Module) Attr(n string) (starlark.Value, error) { return star.Attr(f, n, builtins, properties) }
func (f *Module) AttrNames() []string                   { return star.AttrNames(builtins, properties) }

var properties = map[string]star.PropertyFactory{}

var builtins = map[string]*starlark.Builtin{
	"seed":    starlark.NewBuiltin("seed", seedFn),
	"randint": starlark.NewBuiltin("randint", randIntFn),
	"random":  starlark.NewBuiltin("random", randFn),
}

// seedFn is used to seed the random number generator
// Arguments:
//   - seed: the seed value
//
// Return: None
func seedFn(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var seed int64
	if err := starlark.UnpackArgs("int", args, kwargs, "seed", &seed); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	// create a new random source with the seed
	r := rand.New(rand.NewSource(int64(seed)))

	// store it in thread local storage.
	// If see was called before and a random source was stored in the thread local, it will be replaced
	t.SetLocal(threadLocalSeededRandInstanceKey, r)

	return starlark.None, nil
}

// randIntFn generates a random integer between the specified range
// Arguments:
//   - min: the minimum value of the random integer
//   - max: the maximum value of the random integer
//
// Return: The generated random integer
func randIntFn(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var min, max int
	if err := starlark.UnpackArgs("int", args, kwargs, "min", &min, "max", &max); err != nil {
		logger.Error("error", zap.Error(err))
		return nil, err
	}

	var v int
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		r, ok := t.Local(threadLocalSeededRandInstanceKey).(*rand.Rand)
		if ok && r != nil { // seed was called before and a random source was stored in the thread local. Use it
			return r.Intn(max-min+1) + min
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
func randFn(t *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	ctx := cadstar.GetContext(t)
	logger := workflow.GetLogger(ctx)

	var v float64
	err := workflow.SideEffect(ctx, func(ctx workflow.Context) any {
		r, ok := t.Local(threadLocalSeededRandInstanceKey).(*rand.Rand)
		if ok && r != nil { // seed was called before and a random source was stored in the thread local. Use it
			return r.Float64()
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
