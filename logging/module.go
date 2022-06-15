package logging

import (
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/internal/pkg/typex"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/logging/logutil"
)

func GetGlobal(name string, fields ...zap.Field) *ModuleLogger {
	xerror.Assert(name == "", "[name] should not be null")
	return &ModuleLogger{name: name, fields: fields}
}

// ModuleLog 命名的log
func ModuleLog(log *Logger, name string, fields ...zap.Field) *ModuleLogger {
	xerror.Assert(name == "", "[name] should not be null")
	xerror.Assert(log == nil, "[log] should not be nil")
	return &ModuleLogger{log: log.Named(logutil.Names(logkey.Module, name)).With(fields...)}
}

type ModuleLogger struct {
	name   string
	fields []zap.Field
	log    *Logger
	once   sync.Once
}

func (t *ModuleLogger) If(ok bool, logFn func(log *ModuleLogger)) {
	if ok {
		logFn(t)
	}
}

func (t *ModuleLogger) IfDebug(fn func(log *ModuleLogger)) {
	if t.L().Core().Enabled(zap.DebugLevel) {
		fn(t)
	}
}

func (t *ModuleLogger) IfError(fn func(log *ModuleLogger)) {
	if t.L().Core().Enabled(zap.ErrorLevel) {
		fn(t)
	}
}

func (t *ModuleLogger) WithErr(err error, fields ...zap.Field) *zap.Logger {
	if err == nil {
		return t.L()
	}

	return t.L().With(logutil.ErrField(err, fields...)...)
}

func (t *ModuleLogger) WithFunc(fn interface{}) *zap.Logger {
	if fn == nil {
		return t.L()
	}

	return t.L().With(logutil.FuncStack(fn))
}

func (t *ModuleLogger) L() *zap.Logger {
	if t.log != nil {
		return t.log
	}

	if global != nil {
		t.once.Do(func() {
			t.log = global.Named(logutil.Names(logkey.Module, t.name)).With(t.fields...)
		})
		return t.log
	}

	return zap.L()
}

func (t *ModuleLogger) S() *zap.SugaredLogger {
	return t.L().Sugar()
}

func (t *ModuleLogger) With(args typex.M) *zap.Logger {
	if args == nil || len(args) == 0 {
		return t.L()
	}

	var fields = make([]zap.Field, 0, len(args))
	for k, v := range args {
		fields = append(fields, zap.Any(k, v))
	}

	return t.L().With(fields...)
}

func (t *ModuleLogger) Depth(depth ...int) *zap.Logger {
	if len(depth) == 0 || depth[0] == 0 {
		return t.L()
	}

	return t.L().WithOptions(zap.AddCallerSkip(depth[0]))
}
