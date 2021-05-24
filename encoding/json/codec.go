package json

import (
	"bytes"

	json "github.com/json-iterator/go"
)

// jsonCodec uses json marshaler and unmarshaler.
type jsonCodec struct{}

func (c jsonCodec) Marshal(v interface{}) ([]byte, error) {
	return c.Encode(v)
}

func (c jsonCodec) Unmarshal(data []byte, v interface{}) error {
	return c.Decode(data, v)
}

func (c jsonCodec) Name() string {
	return Name
}

// Encode encodes an object into slice of bytes.
func (c jsonCodec) Encode(i interface{}) ([]byte, error) {
	return json.Marshal(i)
}

// Decode decodes an object from slice of bytes.
func (c jsonCodec) Decode(data []byte, i interface{}) error {
	d := json.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	return d.Decode(i)
}
