package grpcs

import (
	encoding3 "github.com/pubgo/lava/encoding"
	"google.golang.org/grpc/encoding"
)

func init() {
	// 编码注册
	encoding3.Each(func(_ string, cdc encoding3.Codec) {
		encoding.RegisterCodec(cdc)
	})
}
