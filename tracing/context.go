package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

type spanKey struct{}

// withCtx returns a new context with the provided span.
func withCtx(ctx context.Context, span *Span) context.Context {
	return context.WithValue(ctx, spanKey{}, span)
}

// FromCtx retrieves the current span from the context.
func FromCtx(ctx context.Context) *Span {
	span := ctx.Value(spanKey{})
	if span != nil {
		return span.(*Span)
	}

	span = opentracing.SpanFromContext(ctx)
	if span != nil {
		return NewSpan(span.(opentracing.Span))
	}

	return NewSpan((&opentracing.NoopTracer{}).StartSpan(""))
}

func SpanFromCtx(ctx context.Context, fn func(span *Span)) { fn(FromCtx(ctx)) }
