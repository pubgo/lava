package grpc_entry

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"github.com/pubgo/golug/golug_entry"
	"google.golang.org/grpc"
)

type GrpcOptions struct{}
type GrpcOption func(opts *GrpcOptions)
type GrpcEntry interface {
	golug_entry.Entry
	Register(handler interface{}, opts ...GrpcOption)
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
