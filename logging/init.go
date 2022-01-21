package logger

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logger/logkey"
	"github.com/pubgo/lava/runtime"
)

const name = "logger"

var initialized bool

// Init logger
func Init(opts ...func(cfg *xlog_config.Config)) {
	defer func() {
		// 初始化完成
		initialized = true
	}()

	defer xerror.RespExit("logger init error")

	var cfg = xlog_config.NewProdConfig()
	if runtime.IsDev() || runtime.IsTest() || runtime.IsStag() {
		cfg = xlog_config.NewDevConfig()
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
