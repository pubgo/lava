package golug_entry

import (
	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
)

type GrpcOptions struct{}
type GrpcOption func(opts *GrpcOptions)
type GrpcEntry interface {
	Entry
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

type GrpcRestHandler struct {
	Method        string `json:"method"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	ClientStream  bool   `json:"client_stream"`
	ServerStreams bool   `json:"server_streams"`
}
