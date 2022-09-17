package service

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

type GrpcHandler interface {
	Init()
	Middlewares() []Middleware
	ServiceDesc() grpc.ServiceDesc
	TwirpHandler(opts ...interface{}) http.Handler
}

type HttpRouter interface {
	Init()
	BasePrefix() string
	Middlewares() []Middleware
	Router(app *fiber.App)
}
