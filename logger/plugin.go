package logger

import (
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runenv"
)

const name = "logger"

func init() {
	plugin.Register(&plugin.Base{
		Name: name,
		OnInit: func(ent plugin.Entry) {
			var cfg = xlog_config.NewProdConfig()
			if runenv.IsDev() || runenv.IsTest() {
				cfg = xlog_config.NewDevConfig()
				cfg.EncoderConfig.EncodeCaller = "full"
			}

			_ = config.Decode(name, &cfg)
			cfg.Level = runenv.Level
			cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

			// 全局log设置
			var log = xerror.PanicErr(cfg.Build()).(*zap.Logger)
			Init(log)
		},
	})
}
