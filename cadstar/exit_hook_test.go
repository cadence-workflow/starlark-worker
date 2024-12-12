package cadstar

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func TestExitHooks_RegisterAndUnregister(t *testing.T) {
	hooks := &ExitHooks{}

	// Define a dummy starlark function
	dummyFunc := starlark.NewBuiltin("dummy", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, nil
	})

	// Registerying nil should be no-op
	hooks.Register(nil, nil, nil)
	require.Len(t, hooks.hooks, 0)

	// Register the dummy function
	hooks.Register(dummyFunc, nil, nil)
	require.Len(t, hooks.hooks, 1)

	// Unregister the dummy function
	hooks.Unregister(dummyFunc)
	require.Len(t, hooks.hooks, 0)
}

func TestExitHooks_Run(t *testing.T) {
	hooks := &ExitHooks{}

	// Define a dummy starlark function that returns an error
	dummyFuncWithError := starlark.NewBuiltin("dummy_with_error", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, errors.New("dummy error")
	})

	// Define a dummy starlark function that does not return an error
	dummyFunc := starlark.NewBuiltin("dummy", func(thread *starlark.Thread, fn *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
		return starlark.None, nil
	})

	// Register the functions
	hooks.Register(dummyFuncWithError, nil, nil)
	hooks.Register(dummyFunc, nil, nil)

	// Run the hooks
	err := hooks.Run(&starlark.Thread{})
	require.Error(t, err)
	require.Contains(t, err.Error(), "dummy error")
}
