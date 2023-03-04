package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/merge"
)

func New(c *Cfg) opentracing.Tracer {
	cfg := merge.Struct(DefaultCfg(), c).Unwrap()
	assert.Must(cfg.Build())
	return opentracing.GlobalTracer()
}
