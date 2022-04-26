package resource

import (
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"strings"
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/pkg/typex"
)

const Name = "resource"

var resourceList typex.RwMap
var logs = logging.Component(Name)
var mu sync.Mutex

// Remove 删除资源
func Remove(kind string, name string) {
	logs.S().Infow("resource delete", "kind", kind, "name", name)
	check(kind, name)
	resourceList.Del(join(kind, name))
}

// Has 检查资源是否存在
func Has(kind string, name string) bool {
	check(kind, name)
	return resourceList.Has(join(kind, name))
}

func join(names ...string) string {
	return strings.Join(names, "-")
}

func check(kind string, name string) {
	xerror.Assert(kind == "" || name == "", "resource: kind(%s) and name(%s) should not be null", kind, name)
}

func defaultDi(kind string) func(obj inject.Object, field inject.Field) (interface{}, bool) {
	return func(obj inject.Object, field inject.Field) (interface{}, bool) {
		var name = field.Name()
		check(kind, name)

		val, ok := resourceList.Load(join(kind, name))
		if !ok || val == nil {
			logs.L().Error("get object failed",
				zap.String("object", obj.Name()),
				zap.String("type", field.Type()),
				zap.String(logkey.Kind, kind),
				zap.String("name", field.Name()))
			return nil, false
		}
		return val, true
	}
}
