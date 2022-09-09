package service

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"reflect"

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

func Wrap[Request any, Response any](handle func(ctx context.Context, req Request) (rsp Response, err error)) func(ctx *fiber.Ctx) error {
	var handleVal = reflect.ValueOf(handle)
	return func(ctx *fiber.Ctx) error {
		s, err := handle(ctx.Context(), nil)
		return ctx.JSON(s)
	}
}
