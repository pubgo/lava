package logutil

import (
	"github.com/kr/pretty"
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Enabled(lvl zapcore.Level, loggers ...*zap.Logger) bool {
	var log = zap.L()
	if len(loggers) > 0 {
		log = loggers[0]
	}
	return log.Core().Enabled(lvl)
}

func OkOrErr(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	log.Info(msg)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg + " ok")
		return
	}

	log.Error(msg+" error", ErrField(err)...)
	panic(err)
}

func OkOrPanic(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	log.Info(msg)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg + " ok")
		return
	}

	log.Error(msg+" error", ErrField(err)...)
}

func LogOrErr(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg)
		return
	}

	log.Error(msg, ErrField(err)...)
}

func ErrRecord(log *zap.Logger, err error, fieldHandle ...func() Fields) bool {
	if err == nil {
		return false
	}

	var fields []zap.Field
	if len(fieldHandle) > 0 {
		fields = fieldHandle[0]()
	}

	log.WithOptions(zap.AddCallerSkip(1)).With(fields...).Error(err.Error(), ErrField(err)...)
	return true
}

func LogOrPanic(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg)
		return
	}

	log.Error(msg, ErrField(err)...)
	panic(err)
}

func ErrTry(log *zap.Logger, fn func(), fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	var err error
	xerror.TryWith(&err, fn)

	if err == nil {
		return
	}

	log.Error("panic catch", ErrField(err)...)
}

func Pretty(a ...interface{}) {
	zap.L().WithOptions(zap.AddCallerSkip(1)).Info("\n" + pretty.Sprint(a...))
}

func ColorPretty(a ...interface{}) {
	zap.L().WithOptions(zap.AddCallerSkip(1)).Info(string(q.Sq(a...)))
}
