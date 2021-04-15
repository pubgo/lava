package json

import (
	"github.com/pubgo/lug/encoding"
)

var Name = "json"

func init() {
	encoding.Register(Name, jsonCodec{})
}
