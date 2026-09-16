package gateway

import (
	"context"

	"google.golang.org/grpc/metadata"
)

// RPCHandler is the continuation of an RPC middleware chain. It runs the full
// Dispatch call (unary or any streaming mode) after optional request pre-read.
type RPCHandler func(ctx context.Context) (header, trailer metadata.MD, err error)

// RPCMiddleware wraps an entire gateway RPC (not just NewStream/Invoke).
// Use this for cross-cutting logic that must observe the full stream lifetime
// for both in-process and proxy backends.
type RPCMiddleware func(ctx context.Context, op *Operation, next RPCHandler) (header, trailer metadata.MD, err error)

type rpcIncomingPayloadKey struct{}

// IncomingPayload returns the pre-read unary/server-stream request message
// placed on the context by Mux.DispatchFrontend, if any.
func IncomingPayload(ctx context.Context) any {
	return ctx.Value(rpcIncomingPayloadKey{})
}

// UseRPCMiddleware appends middleware around Mux.Dispatch / DispatchFrontend.
func (m *Mux) UseRPCMiddleware(mw RPCMiddleware) {
	if mw == nil {
		return
	}
	m.rpcMiddleware = append(m.rpcMiddleware, mw)
}

func (m *Mux) runRPC(ctx context.Context, op *Operation, next RPCHandler) (metadata.MD, metadata.MD, error) {
	h := next
	for i := len(m.rpcMiddleware) - 1; i >= 0; i-- {
		mw := m.rpcMiddleware[i]
		inner := h
		h = func(ctx context.Context) (metadata.MD, metadata.MD, error) {
			return mw(ctx, op, inner)
		}
	}
	return h(ctx)
}
