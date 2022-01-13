package logz

import (
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logger/logutil"
)

func Component(name string) *Logger {
	xerror.Assert(name == "", "[name] should not be null")
	return &Logger{name: name}
}

type Logger struct {
	name string
}

func (t *Logger) With(args ...zap.Field) *zap.Logger {
	return getName(t.name).With(args...)
}

func (t *Logger) OkOrErr(msg string, fn func() error, fields ...zap.Field) {
	var log = t.Depth(1)
	log = log.With(fields...)

	log.Info(msg)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg + " ok")
		return
	}

	log.Error(msg+" error", logutil.WithErr(err)...)
	panic(err)
}

func (t *Logger) OkOrPanic(msg string, fn func() error, fields ...zap.Field) {
	var log = t.Depth(1)
	log = log.With(fields...)

	log.Info(msg)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg + " ok")
		return
	}

	log.Error(msg+" error", logutil.WithErr(err)...)
}

func (t *Logger) LogOrErr(msg string, fn func() error, fields ...zap.Field) {
	var log = t.Depth(1)
	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg, fields...)
		return
	}

	log.Error(msg, logutil.WithErr(err, fields...)...)
}

func (t *Logger) LogOrPanic(msg string, fn func() error, fields ...zap.Field) {
	var log = t.Depth(1)
	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg, fields...)
		return
	}

	log.Error(msg, logutil.WithErr(err, fields...)...)
	panic(err)
}

func (t *Logger) WithErr(err error, fields ...zap.Field) *zap.Logger {
	if err == nil {
		return discard
	}

	return t.With(logutil.WithErr(err, fields...)...)
}

func (t *Logger) Depth(depth ...int) *zap.Logger {
	if len(depth) > 0 {
		return getName(t.name).WithOptions(zap.AddCallerSkip(depth[0]))
	}
	return getName(t.name)
}

func (t *Logger) Infof(template string, args ...interface{}) {
	getNextName(t.name).Infof(template, args...)
}

func (t *Logger) Info(args ...interface{}) {
	getNextName(t.name).Info(args...)
}

func (t *Logger) Infow(msg string, keysAndValues ...interface{}) {
	getNextName(t.name).Infow(msg, keysAndValues...)
}

func (t *Logger) Errorf(template string, args ...interface{}) {
	getNextName(t.name).Errorf(template, args...)
}

func (t *Logger) Error(args ...interface{}) {
	getNextName(t.name).Error(args...)
}

func (t *Logger) Errorw(msg string, keysAndValues ...interface{}) {
	getNextName(t.name).Errorw(msg, keysAndValues...)
}

func (t *Logger) Warnf(template string, args ...interface{}) {
	getNextName(t.name).Warnf(template, args...)
}

func (t *Logger) Warn(args ...interface{}) {
	getNextName(t.name).Warn(args...)
}

func (t *Logger) Warnw(msg string, keysAndValues ...interface{}) {
	getNextName(t.name).Warnw(msg, keysAndValues...)
}

func (t *Logger) TryWith(fn func()) *zap.SugaredLogger {
	var err error
	xerror.TryWith(&err, fn)
	if err == nil {
		return discard.Sugar()
	}

	return debugLog.Named(t.name).With(logutil.WithErr(err, logutil.FuncStack(fn))...).Sugar()
}

func getName(name string) *zap.Logger {
	if val, ok := loggerMap.Load(name); ok {
		return val.(*zap.Logger)
	}

	var l = debugLog.Named(name)
	loggerMap.Store(name, l)
	return l
}

func getNextName(name string) *zap.SugaredLogger {
	if val, ok := loggerNextMap.Load(name); ok {
		return val.(*zap.SugaredLogger)
	}

	var l = debugNext.Named(name).Desugar().Sugar()
	loggerNextMap.Store(name, l)
	return l
}
