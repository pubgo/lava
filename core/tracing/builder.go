package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/funk/assert"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/merge"
)

func New(c *Cfg, log *logging.Logger) opentracing.Tracer {
	cfg := merge.Struct(DefaultCfg(), c).Unwrap()
	assert.Must(cfg.Build())
	return opentracing.GlobalTracer()
}
