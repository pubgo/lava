package logger

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/runenv"
)

const name = "logger"

// Init logger
func Init(opts ...func(cfg *xlog_config.Config)) {
	defer xerror.RespExit()

	var cfg = xlog_config.NewProdConfig()
	if runenv.IsDev() || runenv.IsTest() || runenv.IsStag() {
		cfg = xlog_config.NewDevConfig()
		cfg.EncoderConfig.EncodeCaller = "full"
	}

	cfg.Level = runenv.Level
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	if len(opts) > 0 && opts[0] != nil {
		opts[0](&cfg)
	}

	// 全局log设置
	var log = cfg.Build(runenv.Project)
	if runenv.Namespace != "" {
		log = log.With(zap.String("env", runenv.Namespace))
	}
	log = log.With(zap.String("project", runenv.Project))

	// 业务日志
	appLog := log.Named(runenv.Name()).With(zap.Namespace("fields"))

	globalLog = appLog.Sugar()
	globalNext = appLog.WithOptions(zap.AddCallerSkip(1)).Sugar()

	// 替换全局zap
	zap.ReplaceGlobals(appLog)

	// 用于logz触发
	xerror.Exit(dix.ProviderNs("lava", log))
}
