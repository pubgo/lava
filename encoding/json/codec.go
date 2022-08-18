package json

import (
	json "github.com/json-iterator/go"
	"github.com/pubgo/lava/encoding"
)

var (
	Name = "json"
	std  = json.Config{
		EscapeHTML:             true,
		UseNumber:              true,
		ValidateJsonRawMessage: true,
	}.Froze()
)

func init() {
	encoding.Register(Name, &jsonCodec{})
}

// jsonCodec uses json marshaler and unmarshaler.
type jsonCodec struct{}

func (c *jsonCodec) Marshal(v interface{}) ([]byte, error)      { return std.Marshal(v) }
func (c *jsonCodec) Unmarshal(data []byte, v interface{}) error { return std.Unmarshal(data, v) }
func (c *jsonCodec) Name() string                               { return Name }
func (c *jsonCodec) Encode(i interface{}) ([]byte, error)       { return std.Marshal(i) }
func (c *jsonCodec) Decode(data []byte, v interface{}) error    { return std.Unmarshal(data, v) }
