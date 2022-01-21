package logger

import (
	"sync"

	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog/xlog_config"
	"go.uber.org/zap"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logger/logkey"
	"github.com/pubgo/lava/logger/logutil"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
)

// 默认log
var componentLog = func() *zap.Logger {
	defer xerror.RespExit()
	cfg := xlog_config.NewDevConfig()
	cfg.EncoderConfig.EncodeCaller = "full"
	cfg.EncoderConfig.EncodeTime = consts.DefaultTimeFormat
	cfg.Rotate = nil
	return cfg.Build(runtime.Name()).Named(logkey.Debug)
}()

var loggerMap sync.Map

// Component 命名的log
func Component(name string, fields ...zap.Field) *namedLogger {
	xerror.Assert(name == "", "[names] should not be null")
	return &namedLogger{name: name, fields: fields}
}

func getName(name string, fields *[]zap.Field) *zap.Logger {
	if val, ok := loggerMap.Load(name); ok {
		return val.(*zap.Logger)
	}

	if !initialized {
		return componentLog.Named(name).With(*fields...)
	}

	var l = componentLog.Named(name).With(*fields...)
	loggerMap.LoadOrStore(name, l)
	return l
}

type namedLogger struct {
	name   string
	fields []zap.Field
}

func (t *namedLogger) WithErr(err error, fields ...zap.Field) *zap.Logger {
	if err == nil {
		return t.L()
	}

	return t.L().With(logutil.WithErr(err, fields...)...)
}

func (t *namedLogger) WithFunc(fn interface{}) *zap.Logger {
	if fn == nil {
		return t.L()
	}

	return t.L().With(logutil.FuncStack(fn))
}

func (t *namedLogger) L() *zap.Logger {
	return getName(t.name, &t.fields)
}

func (t *namedLogger) S() *zap.SugaredLogger {
	return t.L().Sugar()
}

func (t *namedLogger) With(args types.M) *zap.Logger {
	if args == nil || len(args) == 0 {
		return t.L()
	}

	var fields = make([]zap.Field, len(args))
	for k, v := range args {
		fields = append(fields, zap.Any(k, v))
	}

	return t.L().With(fields...)
}

func (t *namedLogger) Depth(depth ...int) *zap.Logger {
	if len(depth) == 0 || depth[0] == 0 {
		return t.L()
	}

	return t.L().WithOptions(zap.AddCallerSkip(depth[0]))
}
