package service

import (
	"github.com/gofiber/fiber/v2"
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
}

type HttpApiRouter interface {
	ApiRouter(app *fiber.App)
}

type HttpDebugRouter interface {
	ApiRouter(app *fiber.App)
}

type HttpInternalRouter interface {
	ApiRouter(app *fiber.App)
}

type HttpAdminRouter interface {
	ApiRouter(app *fiber.App)
}
