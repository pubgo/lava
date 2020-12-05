package golug_log

import (
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/pubgo/golug/golug_entry"
	"github.com/pubgo/golug/golug_plugin"
	"github.com/pubgo/golug/golug_watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"github.com/spf13/pflag"
)

var cfg = xlog_config.NewDevConfig()
var name = "log"

func init() {
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVar(&cfg.Level, "level", cfg.Level, "log level")
		},
		OnInit: func(ent golug_entry.Entry) {
			xerror.Panic(ent.UnWrap(func(entry golug_entry.HttpEntry) {
				entry.Use(logger.New(logger.Config{Format: "${pid} - ${time} ${status} - ${latency} ${method} ${path}\n"}))
			}))

			xerror.Panic(ent.Decode(name, &cfg))
			xerror.Panic(initLog(cfg))
		},
		OnWatch: func(r *golug_watcher.Response) {
			xerror.Panic(r.Decode(&cfg))
			xerror.Panic(initLog(cfg))
		},
	}))
}
