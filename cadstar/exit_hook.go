package cadstar

import (
	"go.starlark.net/starlark"
	"go.uber.org/multierr"
)

type exitHook struct {
	fn     starlark.Callable
	args   starlark.Tuple
	kwargs []starlark.Tuple
}

type ExitHooks struct {
	hooks []*exitHook
}

func (r *ExitHooks) Register(fn starlark.Callable, args starlark.Tuple, kwargs []starlark.Tuple) {
	if fn == nil {
		return
	}

	h := &exitHook{fn: fn, args: args, kwargs: kwargs}
	r.hooks = append(r.hooks, h)
}

func (r *ExitHooks) Unregister(fn starlark.Callable) {
	for i, h := range r.hooks {
		if h.fn == fn {
			r.hooks = append(r.hooks[:i], r.hooks[i+1:]...)
			break
		}
	}
}

func (r *ExitHooks) Run(t *starlark.Thread) error {
	hooks := r.hooks
	var _errors []error
	for i := len(hooks) - 1; i >= 0; i-- {
		h := hooks[i]
		if _, err := starlark.Call(t, h.fn, h.args, h.kwargs); err != nil {
			_errors = append(_errors, err)
		}
	}
	return multierr.Combine(_errors...)
}
