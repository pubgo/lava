package grpc

import (
	grpcMid "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"

	"github.com/pubgo/lug/entry"
)

type options struct{}
type Opt func(opts *options)
type Entry interface {
	entry.Entry
	Init(opts ...grpc.ServerOption)
	Register(handler interface{}, opts ...Opt)
	UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}

type ServerStream = grpcMid.WrappedServerStream

func WrapStream(stream grpc.ServerStream) *grpcMid.WrappedServerStream {
	return grpcMid.WrapServerStream(stream)
}
