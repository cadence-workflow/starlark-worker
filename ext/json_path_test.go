package ext

import (
	"github.com/stretchr/testify/require"
	"testing"
)

func Test_JP(t *testing.T) {

	data := map[string]any{
		"id":      1001,
		"alias":   "foo-alias",
		"removed": true,
		"meta": map[string]any{
			"foo":    "bar",
			"pi":     3.14,
			"vowels": []any{"a", "e", "i", "o", "u", "w", "y"},
			"notes": []any{
				map[string]any{"id": 101, "text": "quick brown fox"},
			},
		},
	}
	t.Run("get-string", func(t *testing.T) {
		res, err := JP[string](data, "alias")
		require.NoError(t, err)
		require.Equal(t, "foo-alias", res)
	})
	t.Run("get-int", func(t *testing.T) {
		res, err := JP[int](data, "id")
		require.NoError(t, err)
		require.Equal(t, 1001, res)
	})
	t.Run("get-bool", func(t *testing.T) {
		res, err := JP[bool](data, "removed")
		require.NoError(t, err)
		require.Equal(t, true, res)
	})
	t.Run("get-float", func(t *testing.T) {
		res, err := JP[float64](data, "meta.pi")
		require.NoError(t, err)
		require.Equal(t, 3.14, res)
	})
	t.Run("not-found-1", func(t *testing.T) {
		_, err := JP[float64](data, "meta.pi2")
		require.Error(t, err)
	})
	t.Run("not-found-2", func(t *testing.T) {
		_, err := JP[float64](data, "meta.pi2.foo")
		require.Error(t, err)
	})
	t.Run("not-found-3", func(t *testing.T) {
		_, err := JP[float64](data, "meta.vowels[100]")
		require.Error(t, err)
	})
	t.Run("get-array-string", func(t *testing.T) {
		res, err := JP[string](data, "meta.vowels[3]")
		require.NoError(t, err)
		require.Equal(t, "o", res)
	})
	t.Run("get-array-string-2", func(t *testing.T) {
		res, err := JP[string](data, "meta.notes[0].text")
		require.NoError(t, err)
		require.Equal(t, "quick brown fox", res)
	})
}
