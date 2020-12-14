package golug_log

import (
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
)

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: Name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVarP(&cfg.Level, "level", cfg.Level, "l", "log level")
		},
		OnInit: func(ent golug_entry.Entry) {
			xerror.Panic(ent.Decode(Name, &cfg))
			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(r *golug_watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	}))
}
