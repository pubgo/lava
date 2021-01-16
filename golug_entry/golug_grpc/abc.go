package golug_grpc

import (
	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/golug/golug_entry"
	"google.golang.org/grpc"
)

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	golug_entry.Entry
	Register(handler interface{}, opts ...Option)
	RegisterUnaryInterceptor(interceptors ...grpc.UnaryServerInterceptor)
	RegisterStreamInterceptor(interceptors ...grpc.StreamServerInterceptor)
}

type WrappedServerStream = grpcMiddleware.WrappedServerStream

func WrapServerStream(stream grpc.ServerStream) *WrappedServerStream {
	return grpcMiddleware.WrapServerStream(stream)
}

type ClientInfo struct {
	Method string
	Conn   *grpc.ClientConn
	Desc   *grpc.StreamDesc
}
