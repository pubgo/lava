package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
)

func GetSpanWithCtx(ctx context.Context) *Span {
	var span = FromCtx(ctx)
	if span == nil {
		span = NewSpan(opentracing.StartSpan("tracing"))
	}
	return span
}

func NewSpan(sp opentracing.Span) *Span {
	return &Span{Span: sp, traceId: GetTraceId(sp.Context())}
}

var _ opentracing.Span = (*Span)(nil)

type Span struct {
	opentracing.Span
	traceId string
}

func (s *Span) SetOperation(name string) *Span {
	s.Span = s.Span.SetOperationName(name)
	return s
}

func (s *Span) CreateChild(name string, opts ...opentracing.StartSpanOption) *Span {
	return &Span{
		traceId: s.traceId,
		Span:    s.Tracer().StartSpan(name, append(opts, opentracing.ChildOf(s.Context()))...),
	}
}

func (s *Span) CreateFollows(name string, opts ...opentracing.StartSpanOption) *Span {
	return &Span{
		traceId: s.traceId,
		Span:    s.Tracer().StartSpan(name, append(opts, opentracing.FollowsFrom(s.Context()))...),
	}
}

func (s *Span) SetBaggage(restrictedKey, value string) *Span {
	s.Span = s.Span.SetBaggageItem(restrictedKey, value)
	return s
}

func (s *Span) TraceID() string { return s.traceId }
func (s *Span) GetHttpHeader() http.Header {
	tracer := opentracing.GlobalTracer()
	header := http.Header{}
	xerror.Panic(tracer.Inject(
		s.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header),
	))
	return header
}
