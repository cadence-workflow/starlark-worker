package star

import (
	"fmt"
	"go.starlark.net/starlark"
)

const CallableObjectType = "callable_object"

var CallableObjectConstructor = starlark.NewBuiltin(CallableObjectType, MakeCallableObject)

// CallableObject is a Callable that supports arbitrary attributes. This functionality is akin to Python functions,
// which can be called and can store custom data in arbitrary attributes.
// See integration_test/testdata/callable_object_test.star for usage examples.
type CallableObject interface {
	starlark.Callable
	starlark.HasSetField
}

type _CallableObject struct {
	src   starlark.Callable
	attrs *starlark.Dict
}

var _ CallableObject = (*_CallableObject)(nil)

func (r *_CallableObject) String() string       { return r.src.String() }
func (r *_CallableObject) Type() string         { return CallableObjectType }
func (r *_CallableObject) Freeze()              { r.src.Freeze(); r.attrs.Freeze() }
func (r *_CallableObject) Truth() starlark.Bool { return r.src.Truth() }
func (r *_CallableObject) Hash() (uint32, error) {
	return 0, fmt.Errorf("unhashable: %s", CallableObjectType)
}

func (r *_CallableObject) Name() string {
	return r.src.Name()
}

func (r *_CallableObject) CallInternal(thread *starlark.Thread, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	return r.src.CallInternal(thread, args, kwargs)
}

func (r *_CallableObject) Attr(name string) (starlark.Value, error) {
	v, found, err := r.attrs.Get(starlark.String(name))
	if err != nil {
		return nil, err
	}
	if !found {
		return nil, nil
	}
	return v, nil
}

func (r *_CallableObject) AttrNames() []string {
	res := make([]string, r.attrs.Len())
	for i, key := range r.attrs.Keys() {
		res[i] = key.(starlark.String).GoString()
	}
	return res
}

func (r *_CallableObject) SetField(name string, value starlark.Value) error {
	return r.attrs.SetKey(starlark.String(name), value)
}

func MakeCallableObject(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	var fn starlark.Callable
	if err := starlark.UnpackArgs(CallableObjectType, args, kwargs, "fn", &fn); err != nil {
		return nil, err
	}
	res := _CallableObject{
		src:   fn,
		attrs: &starlark.Dict{},
	}
	return &res, nil
}
