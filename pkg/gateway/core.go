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
// each gateway frontend (HTTP, gRPC-Web, WebSocket, native gRPC).
type FrontendStream = grpc.ServerStream

// MethodRoute describes a registered gRPC method for external protocol bridges.
type MethodRoute struct {
	FullMethod string
	Operation  *Operation
}

// Operation describes a registered RPC method and its schema.
// It is the single source of truth for dispatch (full method, message types,
// stream mode, rpc meta). HTTP routing indexes into the registry by FullMethod
// via routerTree; it does not duplicate schema fields.
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
	if mth.op != nil {
		return mth.op
	}
	// Lazy build for test helpers that inject methodWrapper without registerRouter.
	mth.op = &Operation{
		FullMethod: mth.grpcFullMethod,
		InputType:  mth.inputType,
		OutputType: mth.outputType,
		StreamDesc: mth.grpcStreamDesc,
		Meta:       mth.meta,
	}
	return mth.op
}

// Dispatcher routes frontend streams to a Backend, handling unary,
// server-streaming, client-streaming, and bidi modes.
type Dispatcher struct{}

func NewDispatcher() *Dispatcher { return &Dispatcher{} }
