package logutil

import (
	"github.com/pubgo/funk/xerr"
	"strings"

	"github.com/pubgo/x/stack"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging/logkey"
)

type Fields = []zap.Field

func Names(names ...string) string {
	return strings.Join(names, ".")
}

func ErrField(err error, fields ...zap.Field) []zap.Field {
	if err == nil {
		return nil
	}

	return append(fields, zap.String(logkey.ErrMsg, err.Error()), zap.String(logkey.ErrDetail, xerr.WrapXErr(err).Stack()))
}

func FnStack(fn interface{}) zap.Field {
	return zap.String(logkey.Stack, stack.Func(fn))
}

type Map map[string]interface{}

func (t Map) Fields() []zap.Field {
	var fields = make([]zap.Field, 0, len(t))
	for k, v := range t {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
