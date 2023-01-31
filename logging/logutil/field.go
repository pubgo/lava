package logutil

import (
	"strings"

	"github.com/pubgo/funk/stack"
	"go.uber.org/zap"

	"github.com/pubgo/lava/logging/logkey"
)

func Names(names ...string) string {
	return strings.Join(names, ".")
}

func FnStack(fn interface{}) zap.Field {
	return zap.String(logkey.Stack, stack.CallerWithFunc(fn).String())
}

type Map map[string]interface{}

func (t Map) Fields() []zap.Field {
	var fields = make([]zap.Field, 0, len(t))
	for k, v := range t {
		fields = append(fields, zap.Any(k, v))
	}
	return fields
}
