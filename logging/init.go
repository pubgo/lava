package logging

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logging/log_config"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/runtime"
)

const Name = "logger"

// 默认log
var componentLog = func() *zap.Logger {
	defer xerror.RespExit()
	var cfg = zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeCaller = zapcore.FullCallerEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(consts.DefaultTimeFormat)
	var log, err = cfg.Build()
	xerror.Panic(err)

	log = log.Named(logkey.Debug)

	// 全局
	zap.ReplaceGlobals(log)
	return log
}()

var initialized bool

// Init logger
func Init(opts ...func(cfg *log_config.Config)) {
	defer func() {
		// 初始化完成
		initialized = true
	}()

	defer xerror.RespExit("logger init error")

	var cfg = log_config.NewProdConfig()
	if runtime.IsDev() || runtime.IsTest() || runtime.IsStag() {
		cfg = log_config.NewDevConfig()
		cfg.EncoderConfig.EncodeCaller = "full"
	}

	cfg.Level = runtime.Level
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat

	if len(opts) > 0 && opts[0] != nil {
		opts[0](&cfg)
	}

	// 全局log设置
	var log = cfg.Build(runtime.Project).With(
		zap.String(logkey.Env, runtime.Mode),
		zap.String(logkey.Project, runtime.Name()),
	)

	if runtime.Namespace != "" {
		log = log.With(zap.String(logkey.Namespace, runtime.Namespace))
	}

	// 基础日志对象, 包含namespace, env, project和项目
	// TODO 版本??
	baseLog := log.With(zap.Namespace(logkey.Fields))

	// 全局log
	globalLog := baseLog.Named(logkey.Service)

	// 替换全局zap全局log
	zap.ReplaceGlobals(globalLog)

	// 组件log
	componentLog = baseLog.Named(logkey.Component)

	// 依赖更新
	xerror.Panic(dix.Provider(&Event{}))
}
