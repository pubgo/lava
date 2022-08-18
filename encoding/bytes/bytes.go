package bytes

import (
	"fmt"

	"github.com/pubgo/lava/encoding"
)

func init() {
	encoding.Register("bytes", &Codec{})
}

type Codec struct {
	data []byte
}

func (c *Codec) Name() string {
	return "bytes"
}

func (c *Codec) Encode(v interface{}) ([]byte, error) {
	switch ve := v.(type) {
	case *[]byte:
		return *ve, nil
	case []byte:
		return ve, nil
	}
	return nil, nil
}

func (c *Codec) Decode(data []byte, ve interface{}) error {
	switch ve := ve.(type) {
	case *[]byte:
		*ve = data
	}
	return nil
}

func (c *Codec) Marshal(v interface{}) ([]byte, error) {
	switch ve := v.(type) {
	case *[]byte:
		return *ve, nil
	case []byte:
		return ve, nil
	}
	return nil, nil
}

func (c *Codec) Unmarshal(data []byte, ve interface{}) error {
	fmt.Println(string(data))
	switch ve := ve.(type) {
	case *[]byte:
		*ve = data
	}
	return nil
}
