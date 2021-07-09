package logger

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/runenv"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"
)

var name = "log"

var cfg = xlog_config.NewProdConfig()

func init() {
	if runenv.IsDev() || runenv.IsTest() {
		cfg = xlog_config.NewDevConfig()
		cfg.EncoderConfig.EncodeCaller = consts.Default
		cfg.EncoderConfig.EncodeTime = "2006-01-02 15:04:05"
	}

	// 全局log设置
	// 默认logger初始化
	xerror.Panic(xlog.SetDefault(xerror.PanicErr(cfg.Build()).(*zap.Logger).Named(runenv.Domain)))

	vars.Watch(name, func() interface{} { return cfg })
}
