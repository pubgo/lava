package logger

import (
	"reflect"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

func Err(err error) zap.Field {
	return zap.Any("err", err)
}

func Name(name string) zap.Field {
	return zap.String("name", name)
}

func Pkg(name string) zap.Field {
	return zap.String("pkg", name)
}

func Id(id string) zap.Field {
	return zap.String("id", id)
}

func UIntPrt(p interface{}) zap.Field {
	return zap.Uintptr("ptr", uintptr(reflect.ValueOf(p).Pointer()))
}

func Try(fn func(), fields ...zap.Field) (gErr error) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	defer xerror.Resp(func(err xerror.XErr) {
		zap.L().Error(err.Error(), append(fields, Err(err), FuncStack(fn))...)
		gErr = err
	})

	fn()
	return
}

func ErrWith(name string, err error) {
	if err == nil {
		return
	}

	zap.L().WithOptions(zap.AddCallerSkip(1)).Error(err.Error(), Err(err), Name(name))
}

func ErrLog(err error, fields ...zap.Field) {
	if err == nil {
		return
	}

	zap.L().WithOptions(zap.AddCallerSkip(1)).Error(err.Error(), append(fields, Err(err))...)
}

func FuncStack(fn interface{}) zap.Field {
	return zap.String("stack", stack.Func(fn))
}

func Logs(fn func(), fields ...zap.Field) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	defer xerror.Resp(func(err xerror.XErr) {
		zap.L().Error(err.Error(), append(fields, Err(err), FuncStack(fn))...)
	})

	fn()
	return
}
