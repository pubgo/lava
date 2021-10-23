package logger

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runenv"
)

const name = "logger"

func init() {
	config.On(func(_ config.Config) {
		defer xerror.RespExit()

		var cfg = xlog_config.NewProdConfig()
		if runenv.IsDev() || runenv.IsTest() {
			cfg = xlog_config.NewDevConfig()
			cfg.EncoderConfig.EncodeCaller = "full"
		}

		_ = config.Decode(name, &cfg)
		cfg.Level = runenv.Level
		cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

		// 全局log设置
		var log = cfg.Build(runenv.Project)
		if env.Namespace == "" {
			log = log.With(zap.String("env", env.Namespace))
		}
		log = log.With(zap.Namespace("fields"))

		// 业务日志
		bisLog := log.Named(runenv.Project)

		// 全局替换
		zap.ReplaceGlobals(bisLog)

		// 依赖注入
		xerror.Exit(dix.Provider(bisLog))
		xerror.Exit(dix.ProviderNs("lava", log))
	})
}
