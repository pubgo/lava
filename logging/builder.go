package logging

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging/log_config"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/runtime"
)

func init() {
	inject.Init(func() {
		New(config.GetCfg())
		inject.Register(fx.Provide(func() *zap.Logger { return zap.L() }))
	})
}

func NewWithCfg(cfg *log_config.Config) {
	cfg.Level = runtime.Level
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	// 全局log设置
	var log = cfg.Build(runtime.Project).With(
		zap.String(logkey.Env, runtime.Mode.String()),
		zap.String(logkey.Hostname, runtime.Hostname),
		zap.String(logkey.Project, runtime.Project),
	)

	if runtime.Namespace != "" {
		log = log.With(zap.String(logkey.Namespace, runtime.Namespace))
	}

	// 基础日志对象, 包含namespace, env, project和项目
	// TODO 版本??
	baseLog := log.With(zap.Namespace(logkey.Fields))

	// 替换zap全局log
	zap.ReplaceGlobals(baseLog)
}

// New logger
func New(c config.Config) {
	defer xerror.RespExit()

	var cfg = log_config.NewProdConfig()
	if runtime.IsDev() || runtime.IsTest() || runtime.IsStag() {
		cfg = log_config.NewDevConfig()
		cfg.EncoderConfig.EncodeCaller = "full"
	}

	xerror.Panic(c.UnmarshalKey(Name, &cfg))
	NewWithCfg(&cfg)
}
