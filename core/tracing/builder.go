package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/internal/pkg/merge"
	"github.com/pubgo/lava/logging"
)

func init() {
	dix.Provider(func(c config.Config, log *logging.Logger) opentracing.Tracer {
		var cfgMap = make(map[string]*Cfg)
		xerror.Panic(c.Decode(Name, &cfgMap))

		for name := range cfgMap {
			var cfg = DefaultCfg()
			if cfgMap[name] != nil {
				xerror.Panic(merge.Struct(&cfg, cfgMap[name]))
			}

			xerror.Panic(cfg.Build())
		}
		return opentracing.GlobalTracer()
	})
}
