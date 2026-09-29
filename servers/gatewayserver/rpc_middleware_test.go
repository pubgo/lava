package gatewayserver

import (
	"context"
	"sync/atomic"
	"testing"

	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/reflect/protoregistry"

	"github.com/pubgo/lava/v2/pkg/gateway"
	"github.com/pubgo/lava/v2/pkg/lava"
)

type countingMiddleware struct {
	hits    *atomic.Int32
	streams *atomic.Int32
}

func (m *countingMiddleware) String() string { return "counting" }

func (m *countingMiddleware) Middleware(next lava.HandlerFunc) lava.HandlerFunc {
	return func(ctx context.Context, req lava.Request) (lava.Response, error) {
		m.hits.Add(1)
		if req.Stream() {
			m.streams.Add(1)
		}
		return next(ctx, req)
	}
}

func TestHandlerRPCMiddle_RunsForUnaryAndStream(t *testing.T) {
	t.Parallel()

	var hits, streams atomic.Int32
	mw := &countingMiddleware{hits: &hits, streams: &streams}
	interceptor := handlerRPCMiddle(map[string][]lava.Middleware{
		"test.v1.Echo": {mw},
	})

	inType, err := protoregistry.GlobalTypes.FindMessageByName("google.protobuf.Empty")
	if err != nil {
		t.Fatalf("type: %v", err)
	}

	streamOp := &gateway.Operation{
		FullMethod: "/test.v1.Echo/Watch",
		InputType:  inType,
		OutputType: inType,
		StreamDesc: &grpc.StreamDesc{ServerStreams: true},
	}
	h, _, err := interceptor(context.Background(), streamOp, func(context.Context) (metadata.MD, metadata.MD, error) {
		return metadata.Pairs("x-from-next", "1"), nil, nil
	})
	if err != nil {
		t.Fatalf("stream rpc mw: %v", err)
	}
	if hits.Load() != 1 || streams.Load() != 1 {
		t.Fatalf("after stream: hits=%d streams=%d", hits.Load(), streams.Load())
	}
	if got := h.Get("x-from-next"); len(got) != 1 || got[0] != "1" {
		t.Fatalf("header=%v", h)
	}

	unaryOp := &gateway.Operation{
		FullMethod: "/test.v1.Echo/Ping",
		InputType:  inType,
		OutputType: inType,
	}
	_, _, err = interceptor(context.Background(), unaryOp, func(context.Context) (metadata.MD, metadata.MD, error) {
		return nil, nil, nil
	})
	if err != nil {
		t.Fatalf("unary rpc mw: %v", err)
	}
	if hits.Load() != 2 || streams.Load() != 1 {
		t.Fatalf("after unary: hits=%d streams=%d", hits.Load(), streams.Load())
	}
}

func TestNewGatewaySurface(t *testing.T) {
	t.Parallel()

	mux := gateway.NewMux()
	g := NewGatewaySurface(mux)
	if g.Mux != mux || g.FiberHandler == nil || g.WebSocketHandler == nil || len(g.GRPCServerOptions) == 0 {
		t.Fatalf("incomplete surface: %+v", g)
	}
}
