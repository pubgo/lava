package logging

import (
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/core/logging/logutil"
	"github.com/pubgo/lava/pkg/typex"
)

var loggerMap sync.Map

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

var Initialized bool

// Component 命名的log
func Component(name string, fields ...zap.Field) *namedLogger {
	xerror.Assert(name == "", "[names] should not be null")
	return &namedLogger{name: logutil.Names(logkey.Component, name), fields: fields}
}

func getName(name string, fields *[]zap.Field) *zap.Logger {
	if val, ok := loggerMap.Load(name); ok {
		return val.(*zap.Logger)
	}

	if !Initialized {
		return componentLog.Named(name).With(*fields...)
	}

	var l = zap.L().Named(name).With(*fields...)
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

	return t.L().With(logutil.ErrField(err, fields...)...)
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

func (t *namedLogger) With(args typex.M) *zap.Logger {
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
