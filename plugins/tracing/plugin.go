package tracing

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/watcher"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func() {
			var cfg = GetDefaultCfg()
			_ = config.Decode(Name, &cfg)
			xerror.Panic(cfg.Build())
		},
		OnWatch: func(name string, resp *watcher.Response) {
			resp.OnPut(func() {
				var cfg = GetDefaultCfg()
				xerror.Panic(types.Decode(resp.Value, &cfg))
				xerror.Panic(cfg.Build())
			})
		},
	})
}
