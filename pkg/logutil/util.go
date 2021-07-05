package logutil

import (
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"go.uber.org/zap"

	"reflect"
)

func Err(err error) zap.Field {
	return zap.Any("err", err)
}

func Name(name string) zap.Field {
	return zap.String("name", name)
}

func Id(id string) zap.Field {
	return zap.String("id", id)
}

func UIntPrt(p interface{}) zap.Field {
	return zap.Uintptr("ptr", uintptr(reflect.ValueOf(p).Pointer()))
}

func Exit(fn func(), fields ...zap.Field)  { log1(xlog.Fatal, fn, fields...) }
func Panic(fn func(), fields ...zap.Field) { log1(xlog.Fatal, fn, fields...) }

func log1(log func(args ...interface{}), fn func(), fields ...zap.Field) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	defer xerror.Resp(func(err xerror.XErr) {
		var params = make([]interface{}, 0, len(fields)+2)
		for i := range fields {
			params = append(params, fields[i])
		}
		log(append(params, zap.Any("err", err), stack.Func(fn))...)
	})

	fn()
}

func Try(fn func(), fields ...zap.Field) (gErr error) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	defer xerror.Resp(func(err xerror.XErr) {
		var params = make([]interface{}, 0, len(fields)+2)
		for i := range fields {
			params = append(params, fields[i])
		}

		xlog.Error(append(params, zap.Any("err", err), stack.Func(fn))...)

		gErr = err
	})

	fn()
	return
}

func Logs(fn func(), fields ...zap.Field) {
	xerror.Assert(fn == nil, "[fn] should not be nil")

	defer xerror.Resp(func(err xerror.XErr) {
		var params = make([]interface{}, 0, len(fields)+2)
		for i := range fields {
			params = append(params, fields[i])
		}

		xlog.Error(append(params, zap.Any("err", err), stack.Func(fn))...)
	})

	fn()
	return
}
