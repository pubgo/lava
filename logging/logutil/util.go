package logutil

import (
	"github.com/kr/pretty"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/xtry"
	"github.com/pubgo/x/q"
	"go.uber.org/zap"
	"strings"
)

func LogOrErr(log *zap.Logger, msg string, fn func() result.Error, fields ...zap.Field) {
	msg = strings.TrimSpace(msg)
	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)

	var err = xtry.TryErr(fn)
	if err.IsNil() {
		log.Info(msg)
	} else {
		log.Error(msg, ErrField(err.Unwrap())...)
	}
}

func OkOrFailed(log *zap.Logger, msg string, fn func() result.Error, fields ...zap.Field) {
	msg = strings.TrimSpace(msg)

	log = log.WithOptions(zap.AddCallerSkip(1)).With(fields...)
	log.Info(msg)

	var err = xtry.TryErr(fn)
	if err.IsNil() {
		log.Info(msg + " ok")
	} else {
		log.Error(msg+" failed", ErrField(err.Unwrap())...)
	}
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
