package logger

import (
	"github.com/pubgo/x/stack"
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

func Id(id string) zap.Field {
	return zap.String("id", id)
}

func FuncStack(fn interface{}) zap.Field {
	return zap.String("stack", stack.Func(fn))
}
