package ext

import (
	"bytes"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestTarDir(t *testing.T) {

	buf := bytes.Buffer{}
	require.NoError(t, DirToTar("./testdata/tar", &buf))

	tarBytes := buf.Bytes()

	actual := map[string][]byte{}
	require.NoError(t, ReadTar(bytes.NewReader(tarBytes), func(path string, content []byte) error {
		actual[path] = content
		return nil
	}))

	expected := map[string][]byte{
		"readme.txt":  []byte("foo bar"),
		"doc/main.md": []byte("quick brown fox"),
	}
	require.Equal(t, expected, actual)
}

func TestTar(t *testing.T) {

	expected := map[string][]byte{
		"app.star": []byte(`
def run(a, b):
    print(a)
    print(b)
    return a + b
`),
	}

	buf := bytes.Buffer{}
	require.NoError(t, WriteTar(expected, &buf))
	tarBytes := buf.Bytes()

	actual := map[string][]byte{}
	err := ReadTar(bytes.NewReader(tarBytes), func(path string, content []byte) error {
		actual[path] = content
		return nil
	})
	require.NoError(t, err)
	require.Equal(t, expected, actual)
}
