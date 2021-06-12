package logutil

import (
	"github.com/pubgo/xlog"
	"go.uber.org/zap"

	"reflect"
)

func Err(err error) xlog.Field {
	return xlog.Any("err", err)
}

func Name(name string) xlog.Field {
	return xlog.String("name", name)
}

func UIntPrt(p interface{}) xlog.Field {
	return zap.Uintptr("ptr", uintptr(reflect.ValueOf(p).Pointer()))
}
