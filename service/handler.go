package service

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/opendoc/opendoc"
	"google.golang.org/grpc"

	_ "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
)

type GrpcHandler interface {
	Init()
	Middlewares() []Middleware
	ServiceDesc() *grpc.ServiceDesc
	Gateway(ctx context.Context, mux *runtime.ServeMux) error
}

type HttpRouter interface {
	Init()
	Middlewares() []Middleware
	Router(app *fiber.App)
	Openapi(swag *opendoc.Swagger)
}
