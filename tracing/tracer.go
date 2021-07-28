package tracing

import (
	"github.com/opentracing/opentracing-go"
)

var GetTraceId = func(span opentracing.SpanContext) string { return "" }

type Tags = opentracing.Tags

type Tracer struct {
	opentracing.Tracer
}

func (t *Tracer) createSpan(name string, opts ...opentracing.StartSpanOption) *Span {
	return NewSpan(t.Tracer.StartSpan(name, opts...))
}

func (t *Tracer) RootSpan(name string, opts ...opentracing.StartSpanOption) *Span {
	return t.createSpan(name, opts...)
}

func (t *Tracer) StartSpan(name string, opts ...opentracing.StartSpanOption) opentracing.Span {
	return t.createSpan(name, opts...)
}
