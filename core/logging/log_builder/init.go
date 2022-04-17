package log_builder

import (
	"go.uber.org/zap"

	logging2 "github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/inject"
)

func init() {
	inject.Register((*logging2.Logger)(nil), func(obj inject.Object, field inject.Field) (interface{}, bool) {
		var name = obj.Name()
		if nm := field.Tag("name"); nm != "" {
			name = nm
		}

		// TODO 更多信息
		return zap.L().Named(name), true
	})
}
