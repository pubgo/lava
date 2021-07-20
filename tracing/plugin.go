package tracing

import (
	"github.com/opentracing/opentracing-go"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/watcher"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name:         Name,
	OnMiddleware: Middleware,
	OnInit: func(ent entry.Entry) {
		var cfg = GetDefaultCfg()
		_ = config.Decode(Name, &cfg)

		var trace = xerror.ExitErr(cfg.Build()).(Tracer)
		opentracing.SetGlobalTracer(&trace)
	},

	OnWatch: func(name string, resp *watcher.Response) {
		resp.OnPut(func() {
			var cfg = GetDefaultCfg()
			xerror.Panic(watcher.Decode(resp.Value, &cfg))

			var trace = xerror.ExitErr(cfg.Build()).(Tracer)
			opentracing.SetGlobalTracer(&trace)
		})
	},
}
