package client

import (
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
)

// middleware for grpc unary calls
var defaultUnaryInterceptor = []grpc.UnaryClientInterceptor{grpc_opentracing.UnaryClientInterceptor()}

// middleware for grpc stream calls
var defaultStreamInterceptor = []grpc.StreamClientInterceptor{grpc_opentracing.StreamClientInterceptor()}

func getDialOption() []grpc.DialOption {
	dialOpts := append(defaultDialOpts,
		grpc.WithChainUnaryInterceptor(defaultUnaryInterceptor...),
		grpc.WithChainStreamInterceptor(defaultStreamInterceptor...))

	return dialOpts
}

func AddUnaryInterceptor(interceptors ...grpc.UnaryClientInterceptor) {
	defaultUnaryInterceptor = append(defaultUnaryInterceptor, interceptors...)
}

func AddStreamInterceptor(interceptors ...grpc.StreamClientInterceptor) {
	defaultStreamInterceptor = append(defaultStreamInterceptor, interceptors...)
}
