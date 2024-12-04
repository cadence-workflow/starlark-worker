package cadstar

import (
	"bufio"
	"bytes"
	"github.com/cadence-workflow/starlark-worker/ext"
	"github.com/cadence-workflow/starlark-worker/star"
	jsoniter "github.com/json-iterator/go"
	"go.starlark.net/starlark"
	"go.uber.org/cadence/encoded"
	"go.uber.org/zap"
	"io"
)

const delimiter byte = '\n'

// DataConverter is a Cadence encoded.DataConverter that supports Starlark types, such as starlark.String, starlark.Int and others.
// Enables passing Starlark values between Cadence workflows and activities.
type DataConverter struct {
	Logger *zap.Logger
}

var _ encoded.DataConverter = (*DataConverter)(nil)

func (s *DataConverter) ToData(values ...any) ([]byte, error) {
	if len(values) == 1 {
		switch v := values[0].(type) {
		case []byte:
			return v, nil
		case starlark.Bytes:
			return []byte(v), nil
		}
	}
	var buf bytes.Buffer
	for _, v := range values {
		b, err := star.Encode(v) // try star encoder
		if _, ok := err.(star.UnsupportedTypeError); ok {
			b, err = jsoniter.Marshal(v) // go encoder fallback
		}
		if err != nil {
			s.Logger.Error("encode-error", ext.ZapError(err)...)
			return nil, err
		}
		buf.Write(b)
		buf.WriteByte(delimiter)
	}
	return buf.Bytes(), nil
}

func (s *DataConverter) FromData(data []byte, to ...any) error {
	if len(to) == 1 {
		switch to := to[0].(type) {
		case *[]byte:
			*to = data
			return nil
		case *starlark.Bytes:
			*to = starlark.Bytes(data)
			return nil
		}
	}
	r := bufio.NewReader(bytes.NewReader(data))

	for i := 0; ; i++ {
		line, err := r.ReadBytes(delimiter)
		var eof bool
		if err == io.EOF {
			eof = true
			err = nil
		}
		if err != nil {
			s.Logger.Error("decode-error", ext.ZapError(err)...)
			return err
		}
		ll := len(line)
		if eof && ll == 0 {
			break
		}
		if line[ll-1] == delimiter {
			line = line[:ll-1]
		}
		out := to[i]
		err = star.Decode(line, out)
		if _, ok := err.(star.UnsupportedTypeError); ok {
			err = jsoniter.Unmarshal(line, out)
		}
		if err != nil {
			s.Logger.Error("decode-error", ext.ZapError(err)...)
			return err
		}
		if eof {
			break
		}
	}
	return nil
}
