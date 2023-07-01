package grpcs

import (
	encoding2 "github.com/pubgo/lava/core/encoding"
	"google.golang.org/grpc/encoding"
)

func init() {
	// 编码注册
	encoding2.Each(func(_ string, cdc encoding2.Codec) {
		encoding.RegisterCodec(cdc)
	})
}
