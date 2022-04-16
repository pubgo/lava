package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// GetFrom retrieves the current span from the context.
func GetFrom(ctx context.Context) *Span {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if sp, ok := span.(*Span); ok {
			return sp
		}
		return NewSpan(span)
	}

	return NewSpan(opentracing.NoopTracer{}.StartSpan("noop"))
}