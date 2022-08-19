package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
)

func New(c *Cfg, log *logging.Logger) opentracing.Tracer {
	var cfg = DefaultCfg()
	assert.Must(merge.Struct(&cfg, c))
	assert.Must(cfg.Build())
	return opentracing.GlobalTracer()
}
