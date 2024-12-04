package star

import (
	"bytes"
	"errors"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTarFS(t *testing.T) {

	files := map[string][]byte{
		"/opt/test/data.txt": []byte("data"),
	}

	bb := bytes.Buffer{}
	require.NoError(t, ext.WriteTar(files, &bb))

	fs, err := NewTarFS(bb.Bytes())
	require.NoError(t, err)

	var data []byte

	data, err = fs.Read("/opt/test/data.txt")
	require.NoError(t, err)
	require.Equal(t, []byte("data"), data)

	data, err = fs.Read("/opt/test/nonexistent_file.txt")
	require.Nil(t, data)
	require.Error(t, err)
	require.True(t, errors.Is(err, ErrNotExist))
}
