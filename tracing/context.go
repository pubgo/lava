package tracing

import (
	"context"

	"github.com/opentracing/opentracing-go"
)

// FromCtx retrieves the current span from the context.
func FromCtx(ctx context.Context) *Span {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if sp, ok := span.(*Span); ok {
			return sp
		}
		return NewSpan(span)
	}

	return nil
}

// FromCtxWith retrieves the current span from the context.
func FromCtxWith(name string, ctx context.Context, cb func(span *Span)) {
	span := opentracing.SpanFromContext(ctx)
	if span != nil {
		if sp, ok := span.(*Span); ok {
			cb(sp)
			return
		}
		cb(NewSpan(span))
		return
	}

	cb(NewSpan(opentracing.StartSpan(name)))
}
