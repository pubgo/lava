package tracing

import (
	"github.com/opentracing/opentracing-go"
)

var GetTraceId = func(span opentracing.SpanContext) string { return "" }

type Tags = opentracing.Tags
