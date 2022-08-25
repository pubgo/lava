package service

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type GatewayHandler func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
type GatewayRegister interface {
	GatewayRegister() GatewayHandler
}

type RegisterServer[T any] func(s grpc.ServiceRegistrar, srv T)
type GrpcRegister[T any] interface {
	GrpcRegister() RegisterServer[T]
}
