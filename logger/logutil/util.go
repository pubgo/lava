package logutil

import (
	"strings"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logger/logkey"
)

func Names(names ...string) string {
	return strings.Join(names, ".")
}

func WithErr(err error, fields ...zap.Field) []zap.Field {
	if err == nil {
		return nil
	}

	return append(fields, zap.String(logkey.Err, err.Error()), zap.Any(logkey.ErrStack, err))
}

func FuncStack(fn interface{}) zap.Field {
	return zap.String(logkey.Stack, stack.Func(fn))
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
