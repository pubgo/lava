package grpc_entry

import (
	"context"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpc_opentracing "github.com/grpc-ecosystem/go-grpc-middleware/tracing/opentracing"
	"github.com/pubgo/golug/golug_config"
	"google.golang.org/grpc"
)

// middleware for grpc unary calls
var defaultUnaryInterceptor = grpc_middleware.ChainUnaryClient(
	grpc_opentracing.UnaryClientInterceptor(),
)

// middleware for grpc stream calls
var defaultStreamInterceptor = grpc_middleware.ChainStreamClient(grpc_opentracing.StreamClientInterceptor())

func init() {
	timeoutCtx, _ := context.WithTimeout(context.Background(), 0)
	_, _ = grpc.DialContext(timeoutCtx, golug_config.Project, grpc.WithChainUnaryInterceptor(defaultUnaryInterceptor),
		grpc.WithChainStreamInterceptor(defaultStreamInterceptor))
}
