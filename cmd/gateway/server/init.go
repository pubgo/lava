package server

import (
	"github.com/pubgo/gateway/codec"
	"google.golang.org/grpc/encoding"
)

func init() {
	encoding.RegisterCodec(codec.NewProxyCodec())
}
