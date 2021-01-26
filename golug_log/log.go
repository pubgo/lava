package golug_log

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/golug_app"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
)

func init() {
	cfg := xlog_config.NewDevConfig()
	cfg.EncoderConfig.EncodeCaller = "full"
	zapL := xerror.PanicErr(xlog_config.NewZapLoggerFromConfig(cfg)).(*zap.Logger)
	log := xlog.New(zapL.WithOptions(xlog.AddCaller(), xlog.AddCallerSkip(1)))

	// 全局log设置
	xerror.Panic(xlog.SetDefault(log.Named(golug_app.Domain, xlog.AddCallerSkip(1))))
}

func initLog(cfg xlog_config.Config) (err error) {
	defer xerror.RespErr(&err)

	zapL := xerror.PanicErr(xlog_config.NewZapLoggerFromConfig(cfg)).(*zap.Logger)
	log := xlog.New(zapL.WithOptions(xlog.AddCaller(), xlog.AddCallerSkip(1)))

	// 全局log设置
	xerror.Panic(xlog.SetDefault(log.Named(golug_app.Domain, xlog.AddCallerSkip(1))))
	// log 变更通知
	xerror.Panic(dix.Dix(log.Named(golug_app.Domain)))
	return nil
}

// Watch
func Watch(fn func(logs xlog.XLog)) {
	defer xerror.RespExit()
	fn(xlog.With())
	xerror.Panic(dix.Dix(fn))
}
