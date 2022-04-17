package log_builder

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	logging2 "github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/log_config"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

// Init logger
func Init(c config.Config) {
	defer func() { logging2.Initialized = true }()

	defer xerror.RespExit("logger init error")

	var cfg = log_config.NewProdConfig()
	if runtime.IsDev() || runtime.IsTest() || runtime.IsStag() {
		cfg = log_config.NewDevConfig()
		cfg.EncoderConfig.EncodeCaller = "full"
	}

	xerror.Panic(c.GetMap(logging2.Name).Decode(&cfg))

	cfg.Level = runtime.Level
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	// 全局log设置
	var log = cfg.Build(runtime.Project).With(
		zap.String(logkey.Env, runtime.Mode.String()),
		zap.String(logkey.Hostname, runtime.Hostname),
		zap.String(logkey.Project, runtime.Name()),
	)

	if runtime.AppID != "" {
		log.With(zap.String(logkey.CommitID, runtime.AppID))
	}

	if runtime.Namespace != "" {
		log = log.With(zap.String(logkey.Namespace, runtime.Namespace))
	}

	// 基础日志对象, 包含namespace, env, project和项目
	// TODO 版本??
	baseLog := log.With(zap.Namespace(logkey.Fields))

	// 替换zap全局log
	zap.ReplaceGlobals(baseLog)
}
