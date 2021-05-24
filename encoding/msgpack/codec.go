package msgpack

import (
	"bytes"

	msgpack "github.com/vmihailenco/msgpack/v5"
)

// msgpackCodec uses messagepack marshaler and unmarshaler.
type msgpackCodec struct{}

func (c msgpackCodec) Marshal(v interface{}) ([]byte, error) {
	return c.Encode(v)
}

func (c msgpackCodec) Unmarshal(data []byte, v interface{}) error {
	return c.Decode(data, v)
}

func (c msgpackCodec) Name() string {
	return Name
}

// Encode encodes an object into slice of bytes.
func (c msgpackCodec) Encode(i interface{}) ([]byte, error) {
	var buf bytes.Buffer
	enc := msgpack.NewEncoder(&buf)
	// enc.UseJSONTag(true)
	err := enc.Encode(i)
	return buf.Bytes(), err
}

// Decode decodes an object from slice of bytes.
func (c msgpackCodec) Decode(data []byte, i interface{}) error {
	dec := msgpack.NewDecoder(bytes.NewReader(data))
	// dec.UseJSONTag(true)
	err := dec.Decode(i)
	return err
}
