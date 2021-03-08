package golug_grpc

import (
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/golug/entry"
	"google.golang.org/grpc"
)

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	entry.Entry
	Register(handler interface{}, opts ...Option)
	InitOpts(opts ...grpc.ServerOption)
	UnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	StreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}

type WrappedServerStream = grpcMiddleware.WrappedServerStream

func WrapServerStream(stream grpc.ServerStream) *WrappedServerStream {
	return grpcMiddleware.WrapServerStream(stream)
}
