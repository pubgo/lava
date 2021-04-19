package tracing

import (
	"github.com/opentracing/opentracing-go"

	"io"
)

type Tracer struct {
	opentracing.Tracer
	io.Closer
}
