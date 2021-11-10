package tracing

import (
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
)

func NewSpan(sp opentracing.Span) *Span {
	return &Span{Span: sp}
}

func StartSpan(operationName string, opts ...opentracing.StartSpanOption) *Span {
	var sp = opentracing.StartSpan(operationName, opts...)
	return &Span{Span: sp}
}

var _ opentracing.Span = (*Span)(nil)

type Span struct {
	opentracing.Span
}

func (s *Span) CreateChild(name string, opts ...opentracing.StartSpanOption) *Span {
	return &Span{Span: s.Tracer().StartSpan(name, append(opts, opentracing.ChildOf(s.Context()))...)}
}

func (s *Span) CreateFollows(name string, opts ...opentracing.StartSpanOption) *Span {
	return &Span{Span: s.Tracer().StartSpan(name, append(opts, opentracing.FollowsFrom(s.Context()))...)}
}

func (s *Span) SetOperationName(name string) opentracing.Span {
	s.Span = s.Span.SetOperationName(name)
	return s
}

func (s *Span) SetTag(key string, value interface{}) opentracing.Span {
	s.Span.SetTag(key, value)
	return s
}

func (s *Span) SetBaggageItem(restrictedKey, value string) opentracing.Span {
	s.Span = s.Span.SetBaggageItem(restrictedKey, value)
	return s
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