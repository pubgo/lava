package golug_tracing

import (
	"net/http"
	"runtime"

	"github.com/opentracing/opentracing-go"
	"github.com/opentracing/opentracing-go/ext"
)

// StartSpan will start a new span with no parent span.
func StartSpan(operationName, method, path string) opentracing.Span {
	return StartSpanWithParent(nil, operationName, method, path)
}

// StartDBSpanWithParent - start a DB operation span
func StartDBSpanWithParent(parent opentracing.SpanContext, operationName, dbInstance, dbType, dbStatement string) opentracing.Span {
	options := []opentracing.StartSpanOption{opentracing.Tag{Key: ext.SpanKindRPCServer.Key, Value: ext.SpanKindRPCServer.Value}}
	if len(dbInstance) > 0 {
		options = append(options, opentracing.Tag{Key: string(ext.DBInstance), Value: dbInstance})
	}
	if len(dbType) > 0 {
		options = append(options, opentracing.Tag{Key: string(ext.DBType), Value: dbType})
	}
	if len(dbStatement) > 0 {
		options = append(options, opentracing.Tag{Key: string(ext.DBStatement), Value: dbStatement})
	}
	if parent != nil {
		options = append(options, opentracing.ChildOf(parent))
	}

	return opentracing.StartSpan(operationName, options...)
}

// StartSpanWithParent will start a new span with a parent span.
// example:
//      span:= StartSpanWithParent(c.Get("tracing-context"),
func StartSpanWithParent(parent opentracing.SpanContext, operationName, method, path string) opentracing.Span {
	options := []opentracing.StartSpanOption{
		opentracing.Tag{Key: ext.SpanKindRPCServer.Key, Value: ext.SpanKindRPCServer.Value},
		opentracing.Tag{Key: string(ext.HTTPMethod), Value: method},
		opentracing.Tag{Key: string(ext.HTTPUrl), Value: path},
		opentracing.Tag{Key: "current-goroutines", Value: runtime.NumGoroutine()},
	}

	if parent != nil {
		options = append(options, opentracing.ChildOf(parent))
	}

	return opentracing.StartSpan(operationName, options...)
}

func InjectTraceID(ctx opentracing.SpanContext, header http.Header) {
	opentracing.GlobalTracer().Inject(
		ctx,
		opentracing.HTTPHeaders,
		opentracing.HTTPHeadersCarrier(header))
}
