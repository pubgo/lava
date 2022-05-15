package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"

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

		inject.NameGroup(Name, name, func(log *logging.Logger) opentracing.Tracer {
			xerror.Exit(cfg.Build())
			return opentracing.GlobalTracer()
		})
	}
}
