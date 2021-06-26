package log

import (
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
)

// 默认logger初始化
func init() {
	cfg := xlog_config.NewDevConfig()
	cfg.EncoderConfig.EncodeCaller = "full"

	// 全局log设置
	xerror.Panic(xlog.SetDefault(xerror.PanicErr(cfg.Build()).(*zap.Logger).Named(runenv.Domain)))
}

func initLog(cfg xlog_config.Config) (err error) {
	defer xerror.RespErr(&err)

	// 全局log设置
	xerror.Panic(xlog.SetDefault(xerror.PanicErr(cfg.Build()).(*zap.Logger).Named(runenv.Domain).Named(runenv.Project)))

	return nil
}
