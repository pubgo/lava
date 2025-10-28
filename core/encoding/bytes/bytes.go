package bytes

import (
	"fmt"

	"github.com/pubgo/lava/v2/core/encoding"
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

func (c *Codec) Encode(v any) ([]byte, error) {
	switch ve := v.(type) {
	case *[]byte:
		return *ve, nil
	case []byte:
		return ve, nil
	}
	return nil, nil
}

func (c *Codec) Decode(data []byte, ve any) error {
	switch ve := ve.(type) {
	case *[]byte:
		*ve = data
	}
	return nil
}

func (c *Codec) Marshal(v any) ([]byte, error) {
	switch ve := v.(type) {
	case *[]byte:
		return *ve, nil
	case []byte:
		return ve, nil
	}
	return nil, nil
}

func (c *Codec) Unmarshal(data []byte, ve any) error {
	fmt.Println(string(data))
	switch ve := ve.(type) {
	case *[]byte:
		*ve = data
	}
	return nil
}
