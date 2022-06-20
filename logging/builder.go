package logging

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/logging/log_config"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/version"
)

func init() {
	dix.Provider(func() ExtLog { return func(log *Logger) {} })
	dix.Provider(func(c config.Config, logs []ExtLog) *Logger {
		var log = New(c)
		for i := range logs {
			logs[i](log)
		}
		return log
	})
}

func NewWithCfg(cfg *log_config.Config) *Logger {
	cfg.Level = runmode.Level
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	// 全局log设置
	var log = cfg.Build(runmode.Project).With(
		zap.String(logkey.Env, runmode.Mode.String()),
		zap.String(logkey.Hostname, runmode.Hostname),
		zap.String(logkey.Project, runmode.Project),
		zap.String(logkey.Version, version.Version),
	)

	if runmode.Namespace != "" {
		log = log.With(zap.String(logkey.Namespace, runmode.Namespace))
	}

	// 基础日志对象, 包含namespace, env, project和项目
	// TODO 版本??
	baseLog := log.With(zap.Namespace(logkey.Fields))

	// 替换zap全局log
	zap.ReplaceGlobals(baseLog)
	global = baseLog
	return baseLog
}

// New logger
func New(c config.Config) *Logger {
	defer funk.RecoverAndExit()

	var cfg = log_config.NewProdConfig()
	if runmode.IsDev() || runmode.IsTest() || runmode.IsStag() {
		cfg = log_config.NewDevConfig()
		cfg.EncoderConfig.EncodeCaller = "full"
	}

	funk.Must(c.UnmarshalKey(Name, &cfg))
	return NewWithCfg(&cfg)
}
