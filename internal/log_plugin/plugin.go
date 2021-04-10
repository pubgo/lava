package log_plugin

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: name,
		OnInit: func(ent interface{}) {
			cfg.Level = config.Level
			if !config.Decode(name, &cfg) {
				return
			}

			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(_ string, r *watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	})
}
