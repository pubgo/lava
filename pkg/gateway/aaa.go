package gateway

import (
	"net/http"

	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type Gateway interface {
	grpc.ClientConnInterface
	SetUnaryInterceptor(interceptor grpc.UnaryServerInterceptor)
	SetStreamInterceptor(interceptor grpc.StreamServerInterceptor)

	SetRequestDecoder(protoreflect.FullName, func(ctx *fiber.Ctx, msg proto.Message) error)
	SetResponseEncoder(protoreflect.FullName, func(ctx *fiber.Ctx, msg proto.Message) error)
	RegisterService(sd *grpc.ServiceDesc, ss interface{})

	Handler(*fiber.Ctx) error
	ServeHTTP(http.ResponseWriter, *http.Request)
}
