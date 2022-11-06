package service

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
)

type GrpcHandler interface {
	Init()
	Middlewares() []Middleware
	ServiceDesc() *grpc.ServiceDesc
	TwirpHandler(...interface{}) http.Handler
}

type GrpcGateway interface {
	GatewayHandler(mux *runtime.ServeMux)
}

func init() {
	runtime.NewServeMux()
}

type HttpRouter interface {
	Init()
	BasePrefix() string
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
