package tracing

import (
	"context"
	"io"

	"github.com/opentracing/opentracing-go"
)

type Tags = opentracing.Tags

type Tracer struct {
	opentracing.Tracer
	io.Closer
}

func (t *Tracer) RootSpan(name string, opts ...opentracing.StartSpanOption) *Span {
	span := new(Span)

	span.Span = t.Tracer.StartSpan(name, opts...)
	span.ctx = opentracing.ContextWithSpan(context.Background(), span.Span)

	return span
}

func (t *Tracer) StartSpan(name string, opts ...opentracing.StartSpanOption) opentracing.Span {
	span := new(Span)

	span.Span = t.Tracer.StartSpan(name, opts...)
	span.ctx = opentracing.ContextWithSpan(context.Background(), span.Span)

	return span
}
