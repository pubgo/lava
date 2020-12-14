package golug_grpc

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/golug/golug_entry"
	"google.golang.org/grpc"
)

type Options struct{}
type Option func(opts *Options)
type Entry interface {
	golug_entry.Entry
	Register(handler interface{}, opts ...Option)
	UnaryServer(interceptors ...grpc.UnaryServerInterceptor)
	StreamServer(interceptors ...grpc.StreamServerInterceptor)
}

type WrappedServerStream = grpc_middleware.WrappedServerStream

func WrapServerStream(stream grpc.ServerStream) *WrappedServerStream {
	return grpc_middleware.WrapServerStream(stream)
}

type ClientInfo struct {
	Method string
	Conn   *grpc.ClientConn
	Desc   *grpc.StreamDesc
}