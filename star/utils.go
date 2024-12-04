package star

import (
	"go.starlark.net/starlark"
	"sort"
)

type PropertyFactory = func(receiver starlark.Value) (starlark.Value, error)

func Attr(
	receiver starlark.Value,
	name string,
	builtins map[string]*starlark.Builtin,
	properties map[string]PropertyFactory,
) (starlark.Value, error) {
	b := builtins[name]
	if b != nil {
		return b.BindReceiver(receiver), nil
	}
	p := properties[name]
	if p != nil {
		return p(receiver)
	}
	return nil, nil
}

func AttrNames(
	builtins map[string]*starlark.Builtin,
	properties map[string]PropertyFactory,
) []string {
	res := make([]string, 0, len(builtins)+len(properties))
	for name := range builtins {
		res = append(res, name)
	}
	for name := range properties {
		res = append(res, name)
	}
	sort.Strings(res)
	return res
}

func Iterate(v starlark.Iterable, handler func(i int, el starlark.Value)) {
	it := v.Iterate()
	defer it.Done()
	var el starlark.Value
	for i := 0; it.Next(&el); i++ {
		handler(i, el)
	}
}

func StringDictToDict(a starlark.StringDict) *starlark.Dict {
	d := starlark.NewDict(len(a))
	for k, v := range a {
		if err := d.SetKey(starlark.String(k), v); err != nil {
			panic(err)
		}
	}
	return d
}

func SetStringKey(d *starlark.Dict, key string, value starlark.Value) {
	if err := d.SetKey(starlark.String(key), value); err != nil {
		panic(err)
	}
}

func StringDictEqualDepth(x, y starlark.StringDict, depth int) (bool, error) {
	if len(x) != len(y) {
		return false, nil
	}
	for key, xv := range x {
		if yv, found := y[key]; !found {
			return false, nil
		} else if eq, err := starlark.EqualDepth(xv, yv, depth-1); err != nil {
			return false, err
		} else if !eq {
			return false, nil
		}
	}
	return true, nil
}

func KeywordsToStringDict(keywords []starlark.Tuple) starlark.StringDict {
	res := make(starlark.StringDict, len(keywords))
	for _, item := range keywords {
		k := item[0].(starlark.String)
		v := item[1]
		res[k.GoString()] = v
	}
	return res
}
