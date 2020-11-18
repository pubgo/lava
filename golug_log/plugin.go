package golug_log

import (
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"github.com/spf13/pflag"
)

var cfg = xlog_config.NewDevConfig()

func init() {
	var name = "log"
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVar(&cfg.Level, "level", cfg.Level, "log level")
		},
		OnInit: func(ent golug_entry.Entry) {
			xerror.Panic(golug_config.Decode(name, &cfg))
			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(r *golug_plugin.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	}))
}
