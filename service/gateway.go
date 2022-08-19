package service

import (
	"context"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type Gateway interface {
	Gateway() func(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}