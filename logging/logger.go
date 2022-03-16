package logging

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/inject"
)

type Logger = zap.Logger

type Fields = []zap.Field

// L global zap log
func L() *zap.Logger {
	return zap.L()
}

// S global zap sugared log
func S() *zap.SugaredLogger {
	return zap.S()
}

type Event struct{}

// On log 依赖注入
func On(fn func(*Event)) {
	xerror.Exit(dix.Provider(fn))
}

func init() {
	inject.Register((*Logger)(nil), func(obj inject.Object, field inject.Field) (interface{}, bool) {
		var name = obj.Name()
		if nm := field.Tag("name"); nm != "" {
			name = nm
		}

		return zap.L().Named(name), true
	})
}
