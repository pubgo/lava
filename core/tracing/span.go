package tracing

import (
	"context"
	"net/http"

	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/assert"
)

func NewSpan(sp opentracing.Span) *Span {
	if sp == nil {
		panic("opentracing.Span is nil")
	}

	if _sp, ok := sp.(*Span); ok {
		return _sp
	}
	return &Span{Span: sp}
}

func StartSpan(operationName string, opts ...opentracing.StartSpanOption) *Span {
	sp := opentracing.StartSpan(operationName, opts...)
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

// SpanID 获取tracerID,spanID
func (s *Span) SpanID() (string, string) {
	return GetSpanID(s.Context())
}

func (s *Span) WithCtx(ctx context.Context) context.Context {
	return opentracing.ContextWithSpan(ctx, s)
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
	assert.Must(tracer.Inject(
		s.Context(),
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header),
	))
	return header
}
