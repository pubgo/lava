package golug_log

import (
	"github.com/pubgo/golug/golug"
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"github.com/spf13/pflag"
)

func init() {
	if golug.IsDev() || golug.IsTest() {
		cfg = xlog_config.NewDevConfig()
	}

	plugin.Register(&plugin.Base{
		Name: name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVarP(&cfg.Level, "level", "l", cfg.Level, "log level")
		},
		OnInit: func(ent interface{}) {
			config.Decode(name, &cfg)
			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(r *watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	})
}
