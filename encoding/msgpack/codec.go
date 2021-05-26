package msgpack

import (
	"github.com/pubgo/lug/encoding"
	msgpack "github.com/vmihailenco/msgpack/v5"
)

var Name = "msgpack"

func init() {
	encoding.Register(Name, msgpackCodec{})
}

// msgpackCodec uses messagepack marshaler and unmarshaler.
type msgpackCodec struct{}

func (c msgpackCodec) Marshal(v interface{}) ([]byte, error)      { return msgpack.Marshal(v) }
func (c msgpackCodec) Unmarshal(data []byte, v interface{}) error { return msgpack.Unmarshal(data, v) }
func (c msgpackCodec) Name() string                               { return Name }
func (c msgpackCodec) Encode(i interface{}) ([]byte, error)       { return msgpack.Marshal(i) }
func (c msgpackCodec) Decode(data []byte, i interface{}) error    { return msgpack.Unmarshal(data, i) }
