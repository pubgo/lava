package gateway

import (
	"google.golang.org/grpc"
	"google.golang.org/protobuf/reflect/protoreflect"

	"github.com/pubgo/lava/v2/pkg/proto/lavapbv1"
)

// Backend is the unified gRPC dispatch target. Mux implements this via
// in-process handlers and optional remote proxy clients.
type Backend = grpc.ClientConnInterface

// FrontendStream is the protocol-neutral request/response surface exposed by
// each gateway frontend (HTTP, gRPC-Web, WebSocket, NATS, etc.).
type FrontendStream = grpc.ServerStream

// Operation describes a registered RPC method and its schema.
type Operation struct {
	FullMethod string
	InputType  protoreflect.MessageType
	OutputType protoreflect.MessageType
	StreamDesc *grpc.StreamDesc // nil for unary RPCs
	Meta       *lavapbv1.RpcMeta
}

func operationFromMethod(mth *methodWrapper) *Operation {
	if mth == nil {
		return nil
	}
	return &Operation{
		FullMethod: mth.grpcFullMethod,
		InputType:  mth.inputType,
		OutputType: mth.outputType,
		StreamDesc: mth.grpcStreamDesc,
		Meta:       mth.meta,
	}
}

// Dispatcher routes frontend streams to a Backend, handling unary,
// server-streaming, client-streaming, and bidi modes.
type Dispatcher struct{}

func NewDispatcher() *Dispatcher { return &Dispatcher{} }
