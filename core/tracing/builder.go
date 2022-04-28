package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/module"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/xerror"
	"go.uber.org/fx"
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

		module.Register(fx.Provide(fx.Annotated{
			Name: module.Name(name),
			Target: func(log *logging.Logger) opentracing.Tracer {
				defer xerror.RespExit()
				xerror.Panic(cfg.Build())
				return opentracing.GlobalTracer()
			},
		}))
	}
}
