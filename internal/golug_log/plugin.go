package golug_log

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/watcher"
	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: name,
		OnInit: func(ent interface{}) {
			cfg.Level = config.Level
			config.Decode(name, &cfg)
			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(_ string, r *watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	})
}
