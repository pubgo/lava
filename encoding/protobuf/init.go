package protobuf

import (
	"github.com/pubgo/lug/encoding"
)

var Name = "protobuf"

func init() {
	encoding.Register(Name, protobufCodec{})
}
