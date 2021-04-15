package msgpack

import (
	"github.com/pubgo/lug/encoding"
)

var Name = "msgpack"

func init() {
	encoding.Register(Name, msgpackCodec{})
}
