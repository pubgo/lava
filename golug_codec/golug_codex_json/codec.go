package golug_codex_json

import (
	"bytes"

	jsoniter "github.com/json-iterator/go"
)

// JSONCodec uses json marshaler and unmarshaler.
type JSONCodec struct{}

// Encode encodes an object into slice of bytes.
func (c JSONCodec) Encode(i interface{}) ([]byte, error) {
	return jsoniter.Marshal(i)
}

// Decode decodes an object from slice of bytes.
func (c JSONCodec) Decode(data []byte, i interface{}) error {
	d := jsoniter.NewDecoder(bytes.NewBuffer(data))
	d.UseNumber()
	return d.Decode(i)
}
