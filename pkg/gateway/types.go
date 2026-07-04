package gateway

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pubgo/lava/v2/pkg/gateway/routertree"
)

type (
	MatchOperation = routertree.MatchOperation
	PathFieldVar   = routertree.PathFieldVar
	RouteOperation = routertree.RouteOperation
	Gateway        interface {
		grpc.ClientConnInterface
		SetUnaryInterceptor(interceptor grpc.UnaryServerInterceptor)
		SetStreamInterceptor(interceptor grpc.StreamServerInterceptor)

		SetRequestDecoder(protoreflect.FullName, func(ctx fiber.Ctx, msg proto.Message) error)
		SetResponseEncoder(protoreflect.FullName, func(ctx fiber.Ctx, msg proto.Message) error)
		RegisterService(sd *grpc.ServiceDesc, ss any)

		GetOperation(operation string) *GrpcMethod
		Handler(fiber.Ctx) error
		ServeHTTP(http.ResponseWriter, *http.Request)
		GetRouteMethods() []RouteOperation
	}
)

// Codec defines the interface used to encode and decode messages.
type Codec interface {
	encoding.Codec
	// MarshalAppend appends the marshaled form of v to b and returns the result.
	MarshalAppend([]byte, any) ([]byte, error)
}

// Compressor is used to compress and decompress messages.
// Based on grpc/encoding.
type Compressor interface {
	encoding.Compressor
}

type (
	MethodHandler  = func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error)
	StreamHandler  = grpc.StreamHandler
	StreamDirector func(ctx context.Context, fullMethodName string) (context.Context, grpc.ClientConnInterface, error)
)
