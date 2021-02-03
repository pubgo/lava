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
	zapL := xerror.PanicErr(xlog_config.NewZapLogger(cfg)).(*zap.Logger)
	log := xlog.New(zapL.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)))

	// 全局log设置
	xerror.Panic(xlog.SetDefault(log.Named(golug_app.Domain, zap.AddCallerSkip(1))))
}

func initLog(cfg xlog_config.Config) (err error) {
	defer xerror.RespErr(&err)

	zapL := xerror.PanicErr(xlog_config.NewZapLogger(cfg)).(*zap.Logger)
	log := xlog.New(zapL.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)))

	// 全局log设置
	xerror.Panic(xlog.SetDefault(log.Named(golug_app.Domain, zap.AddCallerSkip(1))))
	// log 变更通知
	xerror.Panic(dix.Dix(log.Named(golug_app.Domain)))
	return nil
}
