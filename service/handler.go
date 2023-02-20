package service

import (
	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/opendoc/opendoc"
	"google.golang.org/grpc"
)

type GrpcHandler interface {
	Init()
	Middlewares() []Middleware
	ServiceDesc() *grpc.ServiceDesc
}

type HttpRouter interface {
	Init()
	Middlewares() []Middleware
	Router(app *fiber.App)
	Openapi(swag *opendoc.Swagger)
}
