package clientinterceptors

import (
	grpcOpentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"google.golang.org/grpc"
)

func OpenTracingUnary(opts ...grpcOpentracing.Option) grpc.UnaryClientInterceptor {
	return grpcOpentracing.UnaryClientInterceptor(opts...)
}

func OpenTracingStream(opts ...grpcOpentracing.Option) grpc.StreamClientInterceptor {
	return grpcOpentracing.StreamClientInterceptor(opts...)
}
