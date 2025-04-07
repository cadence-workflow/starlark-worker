package temporal

import (
	"testing"

	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.temporal.io/sdk/converter"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
)

// TestStruct test struct.
type TestStruct struct {
	ID int `json:"id"`
}

func TestToPayloads(t *testing.T) {
	testConverter := newTestConverter(t)

	t.Run("encode-go-bytes", func(t *testing.T) {
		payload, err := testConverter.ToPayload([]byte("abc"))
		require.NoError(t, err)

		var out []byte
		err = testConverter.FromPayload(payload, &out)
		require.NoError(t, err)
		require.Equal(t, []byte("abc"), out)
	})

	t.Run("encode-star-bytes", func(t *testing.T) {
		payload, err := testConverter.ToPayload(starlark.Bytes("abc"))
		require.NoError(t, err)

		var out starlark.Bytes
		err = testConverter.FromPayload(payload, &out)
		require.NoError(t, err)
		require.Equal(t, starlark.Bytes("abc"), out)
	})

	t.Run("encode-go-map", func(t *testing.T) {
		payload, err := testConverter.ToPayload(TestStruct{ID: 101})
		require.NoError(t, err)

		var out TestStruct
		err = testConverter.FromPayload(payload, &out)
		require.NoError(t, err)
		require.Equal(t, TestStruct{ID: 101}, out)
	})

	t.Run("encode-several-values", func(t *testing.T) {
		payloads, err := testConverter.ToPayloads(
			starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)},
			true,
		)
		require.NoError(t, err)
		require.Len(t, payloads.Payloads, 2)
	})
}

func TestFromPayloads(t *testing.T) {
	testConverter := newTestConverter(t)

	t.Run("decode-go-bytes", func(t *testing.T) {
		payload, err := testConverter.ToPayload([]byte("abc"))
		require.NoError(t, err)

		var out []byte
		require.NoError(t, testConverter.FromPayload(payload, &out))
		require.Equal(t, []byte("abc"), out)
	})

	t.Run("decode-star-bytes", func(t *testing.T) {
		payload, err := testConverter.ToPayload(starlark.Bytes("abc"))
		require.NoError(t, err)

		var out starlark.Bytes
		require.NoError(t, testConverter.FromPayload(payload, &out))
		require.Equal(t, starlark.Bytes("abc"), out)
	})

	t.Run("decode-go-struct", func(t *testing.T) {
		payload, err := testConverter.ToPayload(TestStruct{ID: 101})
		require.NoError(t, err)

		var out TestStruct
		require.NoError(t, testConverter.FromPayload(payload, &out))
		require.Equal(t, TestStruct{ID: 101}, out)
	})

	t.Run("decode-several", func(t *testing.T) {
		payloads, err := testConverter.ToPayloads(
			starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)},
			true,
			starlark.True,
		)
		require.NoError(t, err)

		var out1 starlark.Tuple
		var out2 any
		var out3 starlark.Value
		require.NoError(t, testConverter.FromPayloads(payloads, &out1, &out2, &out3))

		require.Equal(t, starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)}, out1)
		require.Equal(t, true, out2)
		require.Equal(t, starlark.True, out3)
	})

	t.Run("decode-incompatible-types", func(t *testing.T) {
		payload, err := testConverter.ToPayload(true)
		require.NoError(t, err)

		var out starlark.String
		require.Error(t, testConverter.FromPayload(payload, &out))
	})
}

func newTestConverter(t *testing.T) converter.DataConverter {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	return &DataConverter{Logger: logger}
}
