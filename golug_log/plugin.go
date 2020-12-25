package golug_log

import (
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"github.com/spf13/pflag"
)

func init() {
	if golug_app.IsDev() || golug_app.IsTest() {
		cfg = xlog_config.NewDevConfig()
	}

	golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVarP(&cfg.Level, "level", "l", cfg.Level, "log level")
		},
		OnInit: func(ent golug_entry.Entry) {
			golug_config.Decode(Name, &cfg)
			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(r *golug_watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	})
}
