package tracing_util

import (
	"errors"
	"net/http"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/uber/jaeger-client-go"
)

const (
	// molten兼容
	phpRequestTraceID      = "x-w-traceid"
	phpRequestSpanID       = "x-w-spanid"
	phpRequestParentSpanID = "x-w-parentspanid"
	phpRequestSampleID     = "x-w-sampled"
)

// spanFromPHPRequest 解析php-molten组件链路
func spanFromPHPRequest(req *http.Request) (span jaeger.SpanContext, err error) {
	defer xerror.RecoverErr(&err)

	if req == nil {
		return span, errors.New("context is nil")
	}

	var sampleIDStr = strings.Join(req.Header.Values(phpRequestSampleID), ",")
	var traceIDStr = strings.Join(req.Header.Values(phpRequestTraceID), ",")
	traceID, err := jaeger.TraceIDFromString(traceIDStr)
	xerror.Panic(err)

	var spanIDStr = strings.Join(req.Header.Values(phpRequestSpanID), ",")
	spanID, err := jaeger.SpanIDFromString(spanIDStr)
	xerror.Panic(err)

	var pSpanIDStr = strings.Join(req.Header.Values(phpRequestParentSpanID), ",")
	pSpanID, err := jaeger.SpanIDFromString(pSpanIDStr)
	xerror.Panic(err)

	return jaeger.NewSpanContext(traceID, spanID, pSpanID, sampleIDStr == "", nil), nil
}

// func CreateSpanFromFast(r *fasthttp.Request, name string) opentracing.Span {
// 	tracer := opentracing.GlobalTracer()

// 	var header = make(http.Header)
// 	r.Header.VisitAll(func(key, value []byte) {
// 		header.Add(byteutil.ToStr(key), byteutil.ToStr(value))
// 	})

// 	// If headers contain trace data, create child span from parent; else, create root span
// 	var span opentracing.Span
// 	if tracer != nil {
// 		spanCtx, err := tracer.Extract(opentracing.HTTPHeaders, opentracing.HTTPHeadersCarrier(header))
// 		if err != nil {
// 			span = tracer.StartSpan(name)
// 		} else {
// 			span = tracer.StartSpan(name, ext.RPCServerOption(spanCtx))
// 		}
// 	}

// 	ext.HTTPMethod.Set(span, byteutil.ToStr(r.Header.Method()))
// 	ext.HTTPUrl.Set(span, r.URI().String())

// 	return span // caller must defer span.finish()
// }
