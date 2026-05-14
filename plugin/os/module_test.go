package os

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
)

func newTestModule(t *testing.T, entries map[string]string) *Module {
	t.Helper()
	d := starlark.NewDict(len(entries))
	for k, v := range entries {
		require.NoError(t, d.SetKey(starlark.String(k), starlark.String(v)))
	}
	return &Module{environ: d}
}

func callGetenv(t *testing.T, m *Module, args starlark.Tuple) (starlark.Value, error) {
	t.Helper()
	attr, err := m.Attr("getenv")
	require.NoError(t, err)
	fn, ok := attr.(*starlark.Builtin)
	require.True(t, ok, "expected getenv to be *starlark.Builtin")
	thread := &starlark.Thread{Name: "test"}
	return fn.CallInternal(thread, args, nil)
}

func TestModule_StarlarkValue(t *testing.T) {
	m := newTestModule(t, nil)
	require.Equal(t, pluginID, m.String())
	require.Equal(t, pluginID, m.Type())
	require.True(t, bool(m.Truth()))
	_, err := m.Hash()
	require.Error(t, err)
	require.NotPanics(t, func() { m.Freeze() })
}

func TestModule_AttrNames(t *testing.T) {
	m := newTestModule(t, nil)
	names := m.AttrNames()
	require.Contains(t, names, "getenv")
	require.Contains(t, names, "environ")
}

func TestModule_EnvironAttr(t *testing.T) {
	m := newTestModule(t, map[string]string{"FOO": "bar"})
	v, err := m.Attr("environ")
	require.NoError(t, err)
	d, ok := v.(*starlark.Dict)
	require.True(t, ok)
	got, found, err := d.Get(starlark.String("FOO"))
	require.NoError(t, err)
	require.True(t, found)
	require.Equal(t, starlark.String("bar"), got)
}

func TestModule_UnknownAttr(t *testing.T) {
	m := newTestModule(t, nil)
	v, err := m.Attr("does_not_exist")
	require.NoError(t, err)
	require.Nil(t, v)
}

func TestGetenv_KeyPresent(t *testing.T) {
	m := newTestModule(t, map[string]string{"FOO": "bar"})
	v, err := callGetenv(t, m, starlark.Tuple{starlark.String("FOO")})
	require.NoError(t, err)
	require.Equal(t, starlark.String("bar"), v)
}

func TestGetenv_KeyMissingNoDefault(t *testing.T) {
	m := newTestModule(t, nil)
	v, err := callGetenv(t, m, starlark.Tuple{starlark.String("MISSING")})
	require.NoError(t, err)
	require.Equal(t, starlark.None, v)
}

func TestGetenv_KeyMissingWithDefault(t *testing.T) {
	m := newTestModule(t, nil)
	v, err := callGetenv(t, m, starlark.Tuple{starlark.String("MISSING"), starlark.String("fallback")})
	require.NoError(t, err)
	require.Equal(t, starlark.String("fallback"), v)
}

func TestGetenv_PresentKeyIgnoresDefault(t *testing.T) {
	m := newTestModule(t, map[string]string{"FOO": "bar"})
	v, err := callGetenv(t, m, starlark.Tuple{starlark.String("FOO"), starlark.String("fallback")})
	require.NoError(t, err)
	require.Equal(t, starlark.String("bar"), v)
}

func TestGetenv_MissingKeyArg(t *testing.T) {
	m := newTestModule(t, nil)
	_, err := callGetenv(t, m, nil)
	require.Error(t, err)
}
