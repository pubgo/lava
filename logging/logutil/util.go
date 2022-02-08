package logutil

import (
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func Enabled(lvl zapcore.Level, loggers ...*zap.Logger) (*zap.Logger, bool) {
	var log = zap.L()
	if len(loggers) > 0 {
		log = loggers[0]
	}
	return log, log.Core().Enabled(lvl)
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

	log.Error(msg+" error", WithErr(err)...)
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

	log.Error(msg+" error", WithErr(err)...)
}

func LogOrErr(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg)
		return
	}

	log.Error(msg, WithErr(err)...)
}

func LogOrPanic(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	var err error
	xerror.TryWith(&err, func() { err = fn() })

	if err == nil {
		log.Info(msg)
		return
	}

	log.Error(msg, WithErr(err)...)
	panic(err)
}
