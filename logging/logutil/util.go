package logutil

import (
	"github.com/kr/pretty"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"
	"github.com/pubgo/funk/xtry"
	"github.com/pubgo/x/q"
	"go.uber.org/zap"
)

func OkOrErr(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	log.Info(msg)

	xtry.TryCatch(fn, func(err xerr.XErr) {
		log.Error(msg+" failed", ErrField(err)...)
	})

	log.Info(msg + " ok")
}

func OkOrPanic(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	log.Info(msg)

	xtry.TryCatch(fn, func(err xerr.XErr) {
		log.Error(msg+" error", ErrField(err)...)
		panic(err)
	})

	log.Info(msg + " ok")
}

func LogOrErr(log *zap.Logger, msg string, fn func() error, fields ...zap.Field) {
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	xtry.TryCatch(fn, func(err xerr.XErr) {
		log.Error(msg, ErrField(err)...)
	})

	log.Info(msg)
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

	xtry.TryCatch(fn, func(err xerr.XErr) {
		log.Error(msg, ErrField(err)...)
		panic(err)
	})

	log.Info(msg)
}

func ErrTry(log *zap.Logger, fn func(), fields ...zap.Field) {
	defer recovery.Recovery(func(err xerr.XErr) {
		log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)
		log.Error("panic catch", ErrField(err)...)
	})

	fn()
}

func Pretty(a ...interface{}) {
	zap.L().WithOptions(zap.AddCallerSkip(1)).Info("\n" + pretty.Sprint(a...))
}

func ColorPretty(args ...interface{}) {
	zap.L().WithOptions(zap.AddCallerSkip(1)).Info(string(q.Sq(args...)))
}

func IfDebug(log *zap.Logger, fn func(log *zap.Logger)) {
	if log.Core().Enabled(zap.DebugLevel) {
		fn(log)
	}
}

func IfError(log *zap.Logger, fn func(log *zap.Logger)) {
	if log.Core().Enabled(zap.ErrorLevel) {
		fn(log)
	}
}
