package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/merge"
)

func init() {
	defer xerror.RespExit()
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		var cfg = DefaultCfg()
		if cfgMap[name] != nil {
			xerror.Panic(merge.Struct(&cfg, cfgMap[name]))
		}

		inject.Register(fx.Provide(fx.Annotated{
			Name: inject.Name(name),
			Target: func(log *logging.Logger) opentracing.Tracer {
				xerror.Exit(cfg.Build())
				return opentracing.GlobalTracer()
			},
		}))
	}
}
