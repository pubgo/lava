package tracing

import (
	"context"
)

type spanKey struct{}

// withCtx returns a new context with the provided span.
func withCtx(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey{}, span)
}

// FromCtx retrieves the current span from the context.
func FromCtx(ctx context.Context) *Span {
	spanV := ctx.Value(spanKey{})
	if spanV == nil {
		return nil
	}

	return spanV.(*Span)
}
