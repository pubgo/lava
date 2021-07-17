package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(Register("noop", func(_ map[string]interface{}) (Tracer, error) {
		return Tracer{
			Tracer: opentracing.GlobalTracer(),
		}, nil
	}))
}
