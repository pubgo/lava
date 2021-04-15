package bytes

import (
	"fmt"
	"reflect"
)

// bytesCodec uses raw slice pf bytes and don't encode/decode.
type bytesCodec struct{}

func (c bytesCodec) Name() string {
	return Name
}

// Encode returns raw slice of bytes.
func (c bytesCodec) Encode(i interface{}) ([]byte, error) {
	if data, ok := i.([]byte); ok {
		return data, nil
	}
	if data, ok := i.(*[]byte); ok {
		return *data, nil
	}

	return nil, fmt.Errorf("%T is not a []byte", i)
}

// Decode returns raw slice of bytes.
func (c bytesCodec) Decode(data []byte, i interface{}) error {
	reflect.Indirect(reflect.ValueOf(i)).SetBytes(data)
	return nil
}
