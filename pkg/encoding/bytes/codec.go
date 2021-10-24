package bytes

import (
	"github.com/pubgo/lava/pkg/encoding"
	"reflect"

	"github.com/pubgo/xerror"
)

var Name = "bytes"

func init() {
	encoding.Register(Name, bytesCodec{})
}

// bytesCodec uses raw slice pf bytes and don't encode/decode.
type bytesCodec struct{}

func (c bytesCodec) Marshal(v interface{}) ([]byte, error) {
	ret, err := c.Encode(v)
	if err != nil {
		return nil, xerror.Wrap(err)
	}
	return ret, nil
}

func (c bytesCodec) Unmarshal(data []byte, v interface{}) error {
	if err := c.Decode(data, v); err != nil {
		return xerror.Wrap(err)
	}
	return nil
}

func (c bytesCodec) Name() string { return Name }

// Encode returns raw slice of bytes.
func (c bytesCodec) Encode(i interface{}) ([]byte, error) {
	if data, ok := i.([]byte); ok {
		return data, nil
	}

	if data, ok := i.(*[]byte); ok {
		return *data, nil
	}

	return nil, xerror.Fmt("%T is not a []byte", i)
}

// Decode returns raw slice of bytes.
func (c bytesCodec) Decode(data []byte, i interface{}) error {
	reflect.Indirect(reflect.ValueOf(i)).SetBytes(data)
	return nil
}
