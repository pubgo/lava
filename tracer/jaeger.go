package tracer

import (
	"context"
	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
	"log"
	"net/http"
)

type Span struct {
	Span opentracing.Span
	Ctx  context.Context
}

func NewSpan(ctx context.Context, name string) *Span {
	span := new(Span)
	span.Span, span.Ctx = opentracing.StartSpanFromContext(ctx, name)
	return span
}

func NewRootSpan(name string) *Span {
	return NewSpan(context.Background(), name)
}

func NewSpanByHttpHeader(header *http.Header, name string) *Span {
	traceId := header.Get("uber-trace-id")
	return NewSpanByTraceId(traceId, name)
}

func NewSpanByTraceId(traceId string, name string) *Span {
	carrier := opentracing.HTTPHeadersCarrier{}
	carrier.Set("uber-trace-id", traceId)

	tracer := opentracing.GlobalTracer()
	wireContext, err := tracer.Extract(
		opentracing.HTTPHeaders, carrier)

	if err != nil {
		log.Printf("NewSpanByTraceId err %v\n", err)
		return nil
	}

	span := new(Span)
	span.Span = opentracing.StartSpan(
		name, ext.RPCServerOption(wireContext))

	span.Ctx = opentracing.ContextWithSpan(context.Background(), span.Span)
	return span
}

func (s *Span) SpanContext() opentracing.SpanContext {
	return s.Span.Context()
}

func (s *Span) Context() context.Context {
	return s.Ctx
}

func (s *Span) SetOperationName(name string) *Span {
	s.Span = s.Span.SetOperationName(name)
	return s
}

func (s *Span) LogKV(alternatingKeyValues ...interface{}) {
	s.Span.LogKV(alternatingKeyValues)
}

func (s *Span) SetTag(key string, value interface{}) *Span {
	s.Span = s.Span.SetTag(key, value)
	return s
}

func (s *Span) SetBaggageItem(restrictedKey, value string) *Span {
	s.Span = s.Span.SetBaggageItem(restrictedKey, value)
	return s
}

func (s *Span) Sub(name string) *Span {
	span := new(Span)
	span.Span, span.Ctx = opentracing.StartSpanFromContext(s.Ctx, name)
	return span
}

func (s *Span) GetTraceId() string {
	tracer := opentracing.GlobalTracer()
	header := http.Header{}
	tracer.Inject(s.SpanContext(), opentracing.HTTPHeaders, header)
	return header.Get("uber-trace-id")
}

func (s *Span) GetHttpHeader() http.Header {
	tracer := opentracing.GlobalTracer()
	header := http.Header{}
	tracer.Inject(s.SpanContext(), opentracing.HTTPHeaders, header)
	return header
}

func (s *Span) Finish() {
	s.Span.Finish()
}
