package transport

import "context"

// Transport is transport server.
type Transport interface {
	Endpoint() (string, error)
	Start() error
	Stop() error
}

type transportKey struct{}

// NewCtx returns a new Context that carries value.
func NewCtx(ctx context.Context, tr Transport) context.Context {
	return context.WithValue(ctx, transportKey{}, tr)
}

// FromCtx returns the Transport value stored in ctx, if any.
func FromCtx(ctx context.Context) (tr Transport, ok bool) {
	tr, ok = ctx.Value(transportKey{}).(Transport)
	return
}
