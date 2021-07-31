package logger

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	plugin.Register(&plugin.Base{
		Name:         Name,
		OnMiddleware: Middleware,
		OnInit: func(ent entry.Entry) {
			cfg.Level = runenv.Level
			_ = config.Decode(Name, &cfg)

			var log, err = cfg.Build()
			xerror.Panic(err)

			// 全局log设置
			log = log.Named(runenv.Domain).Named(runenv.Project)
			xerror.Panic(xlog.SetDefault(log))
		},
	})
}
