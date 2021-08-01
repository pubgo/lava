package tracing

import (
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

		xerror.Exit(cfg.Build())
	},

	OnWatch: func(name string, resp *watcher.Response) {
		resp.OnPut(func() {
			var cfg = GetDefaultCfg()
			xerror.Panic(watcher.Decode(resp.Value, &cfg))

			xerror.Exit(cfg.Build())
		})
	},
}
