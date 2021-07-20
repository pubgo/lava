package logger

import (
	"github.com/pubgo/lug/runenv"

	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
)

func updateLog(cfg xlog_config.Config) (err error) {
	return xerror.Wrap(try.Try(func() {
		// 全局log设置
		var log = xerror.PanicErr(cfg.Build()).(*zap.Logger)
		xerror.Panic(xlog.SetDefault(log.Named(runenv.Domain).Named(runenv.Project)))
	}))
}
