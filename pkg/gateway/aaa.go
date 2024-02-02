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

	ServeHTTP(http.ResponseWriter, *http.Request)
	HttpClient() *http.Client

	ServeFast(*fiber.Ctx) error
	FastClient() *fasthttp.Client

	GetApp() *fiber.App
}
