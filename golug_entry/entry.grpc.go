package golug_entry

import (
	"context"

	"google.golang.org/grpc"
)

type GrpcOptions struct{}
type GrpcOption func(opts *GrpcOptions)
type GrpcEntry interface {
	Entry
	Register(handler interface{}, opts ...GrpcOption)
	UnaryServer(interceptors ...UnaryServerInterceptor)
	StreamServer(interceptors ...StreamServerInterceptor)
}

type ClientInfo struct {
	Method string
	Conn   *grpc.ClientConn
	Desc   *grpc.StreamDesc
}

type UnaryServerInterceptor func(ctx context.Context, info *grpc.UnaryServerInfo) context.Context
type StreamServerInterceptor func(ss grpc.ServerStream, info *grpc.StreamServerInfo) context.Context
type UnaryClientInterceptor func(ctx context.Context, info *ClientInfo, opts ...grpc.CallOption)
type StreamClientInterceptor func(ctx context.Context, info *ClientInfo, opts ...grpc.CallOption)

type GrpcRestHandler struct {
	Method        string `json:"method"`
	Name          string `json:"name"`
	Path          string `json:"path"`
	ClientStream  bool   `json:"client_stream"`
	ServerStreams bool   `json:"server_streams"`
}
