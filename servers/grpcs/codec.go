package grpcs

import (
	"google.golang.org/grpc/encoding"

	codec "github.com/pubgo/lava/core/encoding"
)

func init() {
	// 编码注册
	codec.Each(func(_ string, cdc codec.Codec) {
		encoding.RegisterCodec(cdc)
	})
}
