package gateway

import (
	"context"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/emptypb"
)

// The RPC chain owns the full call lifetime, which only Mux.Dispatch can offer:
// Invoke/NewStream hand back before the messages flow. Direct clients get the
// backend chain, and Dispatch must not run the RPC chain twice around it.
func TestRPCMiddleware_BoundaryIsDispatchNotBackendCalls(t *testing.T) {
	var rpcHits, unaryHits, streamHits atomic.Int32

	mux := NewMux()
	mux.UseRPCMiddleware(func(ctx context.Context, op *Operation, next RPCHandler) (metadata.MD, metadata.MD, error) {
		rpcHits.Add(1)
		return next(ctx)
	})
	mux.UseBackendUnaryInterceptor(func(
		ctx context.Context,
		method string,
		req, reply any,
		invoker func(context.Context, string, any, any, ...grpc.CallOption) error,
		opts ...grpc.CallOption,
	) error {
		unaryHits.Add(1)
		return invoker(ctx, method, req, reply, opts...)
	})
	mux.UseBackendStreamInterceptor(func(
		ctx context.Context,
		desc *grpc.StreamDesc,
		method string,
		streamer func(context.Context, *grpc.StreamDesc, string, ...grpc.CallOption) (grpc.ClientStream, error),
		opts ...grpc.CallOption,
	) (grpc.ClientStream, error) {
		streamHits.Add(1)
		return streamer(ctx, desc, method, opts...)
	})

	ctx := context.Background()
	empty := &emptypb.Empty{}
	op := &Operation{
		FullMethod: "/x.Y/Z",
		InputType:  empty.ProtoReflect().Type(),
		OutputType: empty.ProtoReflect().Type(),
	}

	if err := mux.Invoke(ctx, op.FullMethod, empty, empty); err == nil {
		t.Fatal("want unregistered method error from Invoke")
	}
	if _, err := mux.NewStream(ctx, &grpc.StreamDesc{ServerStreams: true}, op.FullMethod); err == nil {
		t.Fatal("want unregistered method error from NewStream")
	}
	if got := rpcHits.Load(); got != 0 {
		t.Fatalf("rpc middleware must not wrap direct backend calls, hits=%d", got)
	}
	if got := unaryHits.Load(); got != 1 {
		t.Fatalf("backend unary interceptor hits=%d want 1", got)
	}
	if got := streamHits.Load(); got != 1 {
		t.Fatalf("backend stream interceptor hits=%d want 1", got)
	}

	if _, _, err := mux.Dispatch(ctx, nil, op, empty); err == nil {
		t.Fatal("want unregistered method error from Dispatch")
	}
	if got := rpcHits.Load(); got != 1 {
		t.Fatalf("rpc middleware must wrap Dispatch exactly once, hits=%d", got)
	}
	if got := unaryHits.Load(); got != 2 {
		t.Fatalf("Dispatch must reuse the backend chain, unary hits=%d want 2", got)
	}
}
