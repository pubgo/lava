package clientinterceptors

import (
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
)

func OpenTracingUnary(opts ...grpc_opentracing.Option) grpc.UnaryClientInterceptor {
	return grpc_opentracing.UnaryClientInterceptor(opts...)
}

func OpenTracingStream(opts ...grpc_opentracing.Option) grpc.StreamClientInterceptor {
	return grpc_opentracing.StreamClientInterceptor(opts...)
}
