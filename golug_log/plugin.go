package golug_log

import (
	"encoding/json"
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/golug_abc"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"github.com/spf13/pflag"
	"go.uber.org/zap"

	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_plugin"
)

func initLog(cfg xlog_config.Config) (err error) {
	defer xerror.RespErr(&err)

	zapL := xerror.PanicErr(xlog_config.NewZapLoggerFromConfig(cfg)).(*zap.Logger)
	log := xlog.New(zapL.WithOptions(xlog.AddCaller(), xlog.AddCallerSkip(1)))
	xerror.Panic(xlog.SetDefault(log.Named(golug_config.Domain, xlog.AddCallerSkip(1))))
	xerror.Panic(dix.Dix(log.Named(golug_config.Domain)))

	trace(cfg)
	return nil
}

func init() {
	var config = xlog_config.NewDevConfig()
	xerror.Exit(golug_plugin.Register(&golug_plugin.Base{
		Name: "log",
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVar(&config.Level, "level", config.Level, "log level")
		},
		OnInit: func(ent golug_abc.Entry) {
			xerror.Panic(golug_config.Decode("log", &config))
			xerror.Panic(initLog(config))
		},
		OnWatch: func(r *golug_plugin.Response) {
			xerror.Panic(json.Unmarshal(r.Value, &config))
			xerror.Panic(initLog(config))
		},
	}))
}
