package golug_log

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
)

func init() {
	zapL := xerror.PanicErr(xlog_config.NewZapLoggerFromConfig(xlog_config.NewDevConfig())).(*zap.Logger)
	log := xlog.New(zapL.WithOptions(xlog.AddCaller(), xlog.AddCallerSkip(1)))

	// 全局log设置
	xerror.Panic(xlog.SetDefault(log.Named(golug_env.Domain, xlog.AddCallerSkip(1))))
}

func initLog(cfg xlog_config.Config) (err error) {
	defer xerror.RespErr(&err)

	zapL := xerror.PanicErr(xlog_config.NewZapLoggerFromConfig(cfg)).(*zap.Logger)
	log := xlog.New(zapL.WithOptions(xlog.AddCaller(), xlog.AddCallerSkip(1)))

	// 全局log设置
	xerror.Panic(xlog.SetDefault(log.Named(golug_env.Domain, xlog.AddCallerSkip(1))))
	// log 变更通知
	xerror.Panic(dix.Dix(log.Named(golug_env.Domain)))
	return nil
}

// getDevLog dev 模式
func getDevLog() xlog.XLog {
	zl, err := xlog_config.NewZapLoggerFromConfig(xlog_config.NewDevConfig())
	if err != nil {
		xerror.Panic(err)
	}

	zl = zl.WithOptions(zap.AddCaller(), zap.AddCallerSkip(1)).Named(golug_env.Project)
	return xlog.New(zl)
}

// Watch
func Watch(fn func(logs xlog.XLog)) {
	defer xerror.RespExit()
	fn(getDevLog())
	xerror.Next().Panic(dix.Dix(fn))
}
