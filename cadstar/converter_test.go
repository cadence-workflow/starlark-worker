package cadstar

import (
	"github.com/stretchr/testify/require"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest"
	"testing"
)

// TestStruct test struct.
type TestStruct struct {
	ID int `json:"id"`
}

// TestToData tests that the converter can encode Go and Starlark values into bytes.
func TestToData(t *testing.T) {
	converter := newTestConverter(t)

	t.Run("encode-go-bytes", func(t *testing.T) {
		// Assert the encoder can encode Go bytes as-is.
		data, err := converter.ToData([]byte("abc"))
		require.NoError(t, err)
		require.Equal(t, []byte("abc"), data)
	})

	t.Run("encode-star-bytes", func(t *testing.T) {
		// Assert the encoder can encode Starlark bytes as-is.
		data, err := converter.ToData(starlark.Bytes("abc"))
		require.NoError(t, err)
		require.Equal(t, []byte("abc"), data)
	})

	t.Run("encode-go-map", func(t *testing.T) {
		// Assert the encoder can encode Go struct as JSON object.
		data, err := converter.ToData(TestStruct{ID: 101})
		require.NoError(t, err)
		require.Equal(t, []byte("{\"id\":101}\n"), data)
	})

	t.Run("encode-several-values", func(t *testing.T) {
		// Assert the encoder can encode several values using the delimiter.
		data, err := converter.ToData(
			starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)},
			true,
		)
		require.NoError(t, err)
		require.Equal(t, []byte("[\"pi\",3.14]\ntrue\n"), data)
	})
}

// TestFromData tests that the converter can decode bytes into Go and Starlark values.
func TestFromData(t *testing.T) {
	converter := newTestConverter(t)

	t.Run("decode-go-bytes", func(t *testing.T) {
		// Assert the encoder can decode into Go bytes as-is.
		var out []byte
		require.NoError(t, converter.FromData([]byte("abc"), &out))
		require.Equal(t, []byte("abc"), out)
	})

	t.Run("decode-star-bytes", func(t *testing.T) {
		// Assert the encoder can decode into Star bytes as-is.
		var out starlark.Bytes
		require.NoError(t, converter.FromData([]byte("abc"), &out))
		require.Equal(t, starlark.Bytes("abc"), out)
	})

	t.Run("decode-go-struct", func(t *testing.T) {
		// Assert the encoder can decode a JSON object into a Go struct.
		var out TestStruct
		require.NoError(t, converter.FromData([]byte("{\"id\":101}\n"), &out))
		require.Equal(t, TestStruct{ID: 101}, out)
	})

	t.Run("decode-no-delimiter", func(t *testing.T) {
		// Assert the encoder can handle the case when the delimiter is missing in the end.
		var out TestStruct
		require.NoError(t, converter.FromData([]byte("{\"id\":101}"), &out))
		require.Equal(t, TestStruct{ID: 101}, out)
	})

	t.Run("decode-several", func(t *testing.T) {
		// Assert the encoder can decode several delimited values.

		var out1 starlark.Tuple
		var out2 any
		var out3 starlark.Value
		require.NoError(t, converter.FromData([]byte("[\"pi\",3.14]\ntrue\ntrue\n"), &out1, &out2, &out3))

		require.Equal(t, starlark.Tuple{starlark.String("pi"), starlark.Float(3.14)}, out1)
		require.Equal(t, true, out2)
		require.Equal(t, starlark.True, out3)
	})

	t.Run("decode-incompatible-types", func(t *testing.T) {
		// Assert the encoder returns an error when trying to decode data into an incompatible type.
		var out starlark.String
		require.Error(t, converter.FromData([]byte("true\n"), &out))
	})
}

func newTestConverter(t *testing.T) encoded.DataConverter {
	logger := zaptest.NewLogger(t, zaptest.Level(zap.InfoLevel))
	return &DataConverter{Logger: logger}
}
