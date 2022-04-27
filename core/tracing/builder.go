package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/pkg/merge"
)

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		if name == consts.KeyDefault {
			name = ""
		}

		var cfg = DefaultCfg()
		xerror.Panic(merge.Struct(&cfg, cfgMap[name]))

		module.Register(fx.Provide(fx.Annotated{
			Name: name,
			Target: func(log *logging.Logger) opentracing.Tracer {
				xerror.Panic(cfg.Build())
				return opentracing.GlobalTracer()
			},
		}))
	}
}
