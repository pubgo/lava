package tracing

import (
	"context"
	"log"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

const TraceId = "trace-id"

func StartSpan(ctx context.Context, name string, opts ...opentracing.StartSpanOption) *Span {
	span := new(Span)
	span.Span, span.ctx = opentracing.StartSpanFromContext(ctx, name, opts...)
	return span
}

func NewSpan(sp opentracing.Span) *Span {
	return &Span{Span: sp}
}

func RootSpan(name string, opts ...opentracing.StartSpanOption) *Span {
	return StartSpan(context.Background(), name, opts...)
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

	return span
}

var _ opentracing.Span = (*Span)(nil)

type Span struct {
	opentracing.Span
	ctx context.Context
}

func (s *Span) Ctx() context.Context { return s.ctx }

func (s *Span) SetOperation(name string) *Span {
	s.Span = s.Span.SetOperationName(name)
	return s
}

func (s *Span) CreateChild(name string, opts ...opentracing.StartSpanOption) *Span {
	return &Span{Span: s.Tracer().StartSpan(name, append(opts, opentracing.ChildOf(s.Context()))...)}
}

func (s *Span) CreateFollows(name string, opts ...opentracing.StartSpanOption) *Span {
	return &Span{Span: s.Tracer().StartSpan(name, append(opts, opentracing.FollowsFrom(s.Context()))...)}
}

func (s *Span) SetBaggage(restrictedKey, value string) *Span {
	s.Span = s.Span.SetBaggageItem(restrictedKey, value)
	return s
}

func (s *Span) GetTraceID() string {
	return GetTraceId(s.Span.Context())
}

func (s *Span) GetHttpHeader() http.Header {
	tracer := opentracing.GlobalTracer()
	header := http.Header{}
	_ = tracer.Inject(s.Context(), opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))
	return header
}

func (s *Span) Finish() {
	s.Span.Finish()
}
