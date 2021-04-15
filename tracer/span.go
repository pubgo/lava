package tracer

import (
	"context"
	"log"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const TraceId = "trace-id"

type Span struct {
	opentracing.Span
	Ctx context.Context
}

func StartSpan(ctx context.Context, name string, opts ...opentracing.StartSpanOption) *Span {
	span := new(Span)
	span.Span, span.Ctx = opentracing.StartSpanFromContext(ctx, name, opts...)
	return span
}

func NewSpan(ctx context.Context, name string, opts ...opentracing.StartSpanOption) *Span {
	span := new(Span)
	span.Span, span.Ctx = opentracing.StartSpanFromContext(ctx, name, opts...)
	return span
}

func NewRootSpan(name string) *Span {
	return NewSpan(context.Background(), name)
}

func NewSpanByHttpHeader(header *http.Header, name string) *Span {
	traceId := header.Get(TraceId)
	return NewSpanByTraceId(traceId, name)
}

func NewSpanByTraceId(traceId string, name string) *Span {
	carrier := opentracing.HTTPHeadersCarrier{}
	carrier.Set(TraceId, traceId)

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

func (s *Span) CreateChild(name string, opts ...opentracing.StartSpanOption) *Span {
	opts = append(opts, opentracing.ChildOf(s.SpanContext()))
	var sp = s.Tracer().StartSpan(name, opts...)
	var ctx = opentracing.ContextWithSpan(s.Context(), sp)
	return &Span{Span: sp, Ctx: ctx}
}

func (s *Span) CreateFollows(name string, opts ...opentracing.StartSpanOption) *Span {
	opts = append(opts, opentracing.FollowsFrom(s.SpanContext()))
	var sp = s.Tracer().StartSpan(name, opts...)
	var ctx = opentracing.ContextWithSpan(s.Context(), sp)
	return &Span{Span: sp, Ctx: ctx}
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
	_ = tracer.Inject(s.SpanContext(), opentracing.HTTPHeaders, header)
	return header.Get(TraceId)
}

func (s *Span) GetHttpHeader() http.Header {
	tracer := opentracing.GlobalTracer()
	header := http.Header{}
	_ = tracer.Inject(s.SpanContext(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))
	return header
}

func (s *Span) Finish() {
	s.Span.Finish()
}
