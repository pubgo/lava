package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
)

const TraceId = "trace-id"

func NewSpan(sp opentracing.Span) *Span {
	return &Span{Span: sp}
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
	xerror.Panic(tracer.Inject(
		s.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header),
	))
	return header
}
