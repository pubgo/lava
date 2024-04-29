package gateway

import (
	"io"
	"net/http"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lava/pkg/gateway/internal/routex"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

type (
	RouteTarget = routex.RouteTarget
	Gateway     interface {
		grpc.ClientConnInterface
		SetUnaryInterceptor(interceptor grpc.UnaryServerInterceptor)
		SetStreamInterceptor(interceptor grpc.StreamServerInterceptor)

		SetRequestDecoder(protoreflect.FullName, func(ctx *fiber.Ctx, msg proto.Message) error)
		SetResponseEncoder(protoreflect.FullName, func(ctx *fiber.Ctx, msg proto.Message) error)
		RegisterService(sd *grpc.ServiceDesc, ss interface{})

		Handler(*fiber.Ctx) error
		ServeHTTP(http.ResponseWriter, *http.Request)
		GetRouteMethods() []*routex.RouteTarget
	}
)

// Codec defines the interface used to encode and decode messages.
type Codec interface {
	encoding.Codec
	// MarshalAppend appends the marshaled form of v to b and returns the result.
	MarshalAppend([]byte, interface{}) ([]byte, error)
}

// StreamCodec is used in streaming RPCs where the message boundaries are
// determined by the codec.
type StreamCodec interface {
	Codec

	// ReadNext returns the size of the next message appended to buf.
	// ReadNext reads from r until either it has read a complete message or
	// encountered an error and returns all the data read from r.
	// The message is contained in dst[:n].
	// Excess data read from r is stored in dst[n:].
	ReadNext(buf []byte, r io.Reader, limit int) (dst []byte, n int, err error)
	// WriteNext writes the message to w with a size aware encoding
	// returning the number of bytes written.
	WriteNext(w io.Writer, src []byte) (n int, err error)
}

// Compressor is used to compress and decompress messages.
// Based on grpc/encoding.
type Compressor interface {
	encoding.Compressor
}
