package gateway

import (
	"context"

	"github.com/pubgo/funk/v2/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var clientStreamDescForProxying = &grpc.StreamDesc{
	ServerStreams: true,
	ClientStreams: true,
}

// callOptionsBackend applies fixed CallOptions on every NewStream/Invoke.
type callOptionsBackend struct {
	cli  grpc.ClientConnInterface
	opts []grpc.CallOption
}

func (b *callOptionsBackend) Invoke(ctx context.Context, method string, args, reply any, opts ...grpc.CallOption) error {
	return b.cli.Invoke(ctx, method, args, reply, append(append([]grpc.CallOption{}, b.opts...), opts...)...)
}

func (b *callOptionsBackend) NewStream(ctx context.Context, desc *grpc.StreamDesc, method string, opts ...grpc.CallOption) (grpc.ClientStream, error) {
	return b.cli.NewStream(ctx, desc, method, append(append([]grpc.CallOption{}, b.opts...), opts...)...)
}

// TransparentHandler returns a gRPC stream handler that proxies an unknown
// method to cli using the shared Dispatcher bidi pump (with backend header
// propagation enabled for remote connections).
func TransparentHandler(cli grpc.ClientConnInterface, inType, outType protoreflect.MessageType, opts ...grpc.CallOption) grpc.StreamHandler {
	if cli == nil {
		panic("gateway: TransparentHandler cli is nil")
	}
	backend := Backend(cli)
	if len(opts) > 0 {
		backend = &callOptionsBackend{cli: cli, opts: opts}
	}
	dispatcher := NewDispatcher()

	return func(_ any, serverStream grpc.ServerStream) error {
		fullMethodName, ok := grpc.MethodFromServerStream(serverStream)
		if !ok {
			return status.Error(codes.Internal, "gateway: method not found in stream context")
		}
		if inType == nil || outType == nil {
			return status.Error(codes.Internal, "gateway: TransparentHandler requires input/output message types")
		}

		op := &Operation{
			FullMethod: fullMethodName,
			InputType:  inType,
			OutputType: outType,
			StreamDesc: clientStreamDescForProxying,
		}
		_, _, err := dispatcher.DispatchFrontend(
			serverStream.Context(),
			backend,
			serverStream,
			op,
			WithPropagateBackendHeaders(),
		)
		if err != nil {
			return errors.WrapCaller(err)
		}
		return nil
	}
}
