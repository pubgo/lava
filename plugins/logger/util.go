package logger

import (
	"time"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

func WithErr(err error, fields ...zap.Field) []zap.Field {
	if err == nil {
		return nil
	}

	return append(fields, zap.String("err", err.Error()), zap.Any("err_stack", err))
}

func Name(name string) zap.Field {
	return zap.String("name", name)
}

func Pkg(name string) zap.Field {
	return zap.String("pkg", name)
}

func Duration(t time.Duration) zap.Field {
	return zap.String("duration", t.String())
}

func Id(id string) zap.Field {
	return zap.String("id", id)
}

func TryWith(fn func(), fields ...zap.Field) *zap.Logger {
	var err error
	xerror.TryWith(&err, fn)
	if err == nil {
		return Discard
	}

	return zap.L().With(WithErr(err, fields...)...)
}

func ErrWith(err error, fields ...zap.Field) *zap.Logger {
	if err == nil {
		return Discard
	}

	return zap.L().With(WithErr(err, fields...)...)
}

func FuncStack(fn interface{}) zap.Field {
	return zap.String("stack", stack.Func(fn))
}
