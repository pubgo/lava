package bytes

import "github.com/pubgo/lug/encoding"

var Name = "bytes"

func init() {
	encoding.Register(Name, bytesCodec{})
}
