package logger

import (
	"github.com/pubgo/lug/runenv"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
)

func initLog(cfg xlog_config.Config) (err error) {
	defer xerror.RespErr(&err)

	// 全局log设置
	var log = xerror.PanicErr(cfg.Build()).(*zap.Logger)
	xerror.Panic(xlog.SetDefault(log.Named(runenv.Domain).Named(runenv.Project)))

	return nil
}
