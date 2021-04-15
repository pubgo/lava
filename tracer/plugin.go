package tracer

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		OnInit: func(ent interface{}) {
			var cfg = GetDefaultCfg()
			_ = config.Decode(Name, &cfg)

			var trace = xerror.PanicErr(cfg.Build()).(opentracing.Tracer)
			opentracing.SetGlobalTracer(trace)
		},

		OnWatch: func(name string, resp *watcher.Response) {
			resp.OnPut(func() {
				var cfg = GetDefaultCfg()
				xerror.Panic(resp.Decode(&cfg))

				var trace = xerror.PanicErr(cfg.Build()).(opentracing.Tracer)
				opentracing.SetGlobalTracer(trace)
			})
		},
	})
}
