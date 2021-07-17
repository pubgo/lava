package logger

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/watcher"

	"github.com/pubgo/xerror"
)

func init() {
	plugin.Register(&plugin.Base{
		Name:         name,
		OnMiddleware: Middleware,
		OnInit: func(ent entry.Entry) {
			cfg.Level = runenv.Level
			_ = config.Decode(name, &cfg)

			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(_ string, r *watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	})
}
