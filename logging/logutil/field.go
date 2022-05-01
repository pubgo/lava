package logutil

import (
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

	return append(fields, zap.String(logkey.Err, err.Error()), zap.Any(logkey.ErrStack, err))
}

func FuncStack(fn interface{}) zap.Field {
	return zap.String(logkey.Stack, stack.Func(fn))
}

type Map map[string]interface{}

func (t Map) Fields() []zap.Field {
	var fields []zap.Field
	for k, v := range t {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
