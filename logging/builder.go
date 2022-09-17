package logging

import (
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/logging/logconfig"
	"github.com/pubgo/lava/logging/logkey"
)

func NewWithCfg(cfg *logconfig.Config) *Logger {
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	// 全局log设置
	var log = cfg.Build(runmode.Project).With(
		zap.String(logkey.Hostname, runmode.Hostname),
		zap.String(logkey.Project, runmode.Project),
		zap.String(logkey.Version, runmode.Version),
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
	defer recovery.Exit()

	var cfg = logconfig.NewProdConfig()

	if runmode.IsDebug {
		cfg.EncoderConfig.EncodeCaller = "full"
	}

	assert.Must(c.UnmarshalKey(Name, &cfg))
	return NewWithCfg(&cfg)
}
