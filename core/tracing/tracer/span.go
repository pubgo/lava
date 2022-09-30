package tracer

import (
	"context"

	"go.opentelemetry.io/otel/trace"
)

// Span warps trace.Span for compatibility and extension.
type Span struct {
	trace.Span
}

// NewSpan creates a span using default tracer.
func NewSpan(ctx context.Context, spanName string, opts ...trace.SpanStartOption) (context.Context, *Span) {
	ctx, span := NewTracer().Start(ctx, spanName, opts...)
	return ctx, &Span{
		Span: span,
	}
}
