package service

import (
	"context"
	"github.com/gofiber/fiber/v2"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type GrpcHandler interface {
	GrpcHandler(reg grpc.ServiceRegistrar)
}

type GrpcGatewayHandler interface {
	GrpcGatewayHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error
}

type HttpRouter interface {
	HttpRouter(app *fiber.App)
}
