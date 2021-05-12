package rpc

import (
	grpcMid "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/lug/entry"
	"google.golang.org/grpc"
)

type Opts struct{}
type Opt func(opts *Opts)
type Entry interface {
	entry.Entry
	Register(handler interface{}, opts ...Opt)
	InitOpts(opts ...grpc.ServerOption)
	UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}

type ServerStream = grpcMid.WrappedServerStream

func WrapStream(stream grpc.ServerStream) *grpcMid.WrappedServerStream {
	return grpcMid.WrapServerStream(stream)
}
