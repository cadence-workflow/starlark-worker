package star

import (
	"encoding/json"
	"fmt"
	"go.starlark.net/starlark"
	"go.starlark.net/syntax"
)

const DataclassType = "dataclass"

const codecAttr = "__codec__"

var DataclassConstructor = starlark.NewBuiltin(DataclassType, MakeDataclass)

type Dataclass interface {
	starlark.Comparable
	starlark.HasSetField
	json.Marshaler
}

type _Dataclass struct {
	attrs starlark.StringDict
}

var _ Dataclass = (*_Dataclass)(nil)

func NewDataclass(attrs starlark.StringDict) Dataclass {
	attrsCopy := starlark.StringDict{}
	for k, v := range attrs {
		attrsCopy[k] = v
	}
	if _, found := attrsCopy[codecAttr]; !found {
		attrsCopy[codecAttr] = starlark.String(DataclassType)
	}
	return &_Dataclass{attrs: attrsCopy}
}

func NewDataclassFromDict(dict *starlark.Dict) Dataclass {
	attrs := KeywordsToStringDict(dict.Items())
	return NewDataclass(attrs)
}

func (r *_Dataclass) String() string                           { return fmt.Sprintf("<%s %s>", DataclassType, r.attrs.String()) }
func (r *_Dataclass) Type() string                             { return DataclassType }
func (r *_Dataclass) Freeze()                                  { r.attrs.Freeze() }
func (r *_Dataclass) Truth() starlark.Bool                     { return len(r.attrs) > 0 }
func (r *_Dataclass) Hash() (uint32, error)                    { return 0, fmt.Errorf("unhashable: %s", DataclassType) }
func (r *_Dataclass) Attr(name string) (starlark.Value, error) { return r.attrs[name], nil }
func (r *_Dataclass) AttrNames() []string                      { return r.attrs.Keys() }

func (r *_Dataclass) CompareSameType(op syntax.Token, y_ starlark.Value, depth int) (bool, error) {
	y := y_.(*_Dataclass)
	switch op {
	case syntax.EQL:
		ok, err := StringDictEqualDepth(r.attrs, y.attrs, depth)
		return ok, err
	case syntax.NEQ:
		ok, err := StringDictEqualDepth(r.attrs, y.attrs, depth)
		return !ok, err
	default:
		return false, fmt.Errorf("%s %s %s not implemented", r.Type(), op, y.Type())
	}
}

func (r *_Dataclass) SetField(name string, val starlark.Value) error {
	if _, found := r.attrs[name]; !found {
		return fmt.Errorf("attribute not found: %s", name)
	}
	r.attrs[name] = val
	return nil
}

func (r *_Dataclass) MarshalJSON() ([]byte, error) {
	if _, found := r.attrs[codecAttr]; !found {
		return nil, fmt.Errorf("invalid state: '%s' attribute not found", codecAttr)
	}
	dict := StringDictToDict(r.attrs)
	return Encode(dict)
}

func MakeDataclass(_ *starlark.Thread, _ *starlark.Builtin, args starlark.Tuple, kwargs []starlark.Tuple) (starlark.Value, error) {
	if args.Len() > 0 {
		return nil, fmt.Errorf("%s() takes no positional arguments", DataclassType)
	}
	attrs := KeywordsToStringDict(kwargs)
	return NewDataclass(attrs), nil
}
