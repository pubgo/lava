package logutil

import (
	"fmt"
	"reflect"

	"github.com/pubgo/xlog"
	"go.uber.org/zap"
)

func ErrWith(err error, msg string, args ...interface{}) {
	if err == nil {
		return
	}

	xlog.Error(fmt.Sprintf(msg, args...), Err(err))
}

func Err(err error) xlog.Field {
	return xlog.Any("err", err)
}

func Name(name string) xlog.Field {
	return xlog.String("name", name)
}

func UIntPrt(p interface{}) xlog.Field {
	return zap.Uintptr("ptr", uintptr(reflect.ValueOf(p).Pointer()))
}
