package logutil

import (
	"reflect"

	"go.uber.org/zap"
)

func Err(err error) zap.Field {
	return zap.Any("err", err)
}

func Name(name string) zap.Field {
	return zap.String("name", name)
}

func UIntPrt(p interface{}) zap.Field {
	return zap.Uintptr("ptr", uintptr(reflect.ValueOf(p).Pointer()))
}
