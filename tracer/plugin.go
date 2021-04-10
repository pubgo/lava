package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		OnInit: func(ent interface{}) {
			var cfg = GetDefaultCfg()
			if !config.Decode(Name, &cfg) {
				return
			}

			driver := cfg.Driver
			xerror.Assert(driver == "", "tracer driver is null")

			fc := Get(driver)
			xerror.Assert(fc == nil, "tracer driver %s not found", driver)

			opentracing.SetGlobalTracer(xerror.PanicErr(fc(config.Map(Name))).(opentracing.Tracer))
		},
	})
}
