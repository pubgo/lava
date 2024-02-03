package gateway

import (
	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp"
	"google.golang.org/grpc"
	"net/http"
)

type Gateway interface {
	grpc.ClientConnInterface
	WithServerUnaryInterceptor(interceptor grpc.UnaryServerInterceptor)
	WithServerStreamInterceptor(interceptor grpc.StreamServerInterceptor)
	RegisterService(sd *grpc.ServiceDesc, ss interface{})

	Handler(*fiber.Ctx) error
	HttpClient() *http.Client
	FastClient() *fasthttp.Client

	App() *fiber.App
}
