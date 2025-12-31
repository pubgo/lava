package grpcs

import (
	codec "github.com/pubgo/lava/v2/core/encoding"
	"google.golang.org/grpc/encoding"
)

func init() {
	// 编码注册
	codec.Each(func(_ string, cdc codec.Codec) {
		encoding.RegisterCodec(cdc)
	})
}
