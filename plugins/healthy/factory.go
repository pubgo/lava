package healthy

import (
	"github.com/pubgo/lava/plugins/healthy/healthy_type"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/pkg/utils"
)

const Name = "health"

var healthList typex.SMap

func Get(names ...string) healthy_type.Handler {
	val, ok := healthList.Load(utils.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(healthy_type.Handler)
}

func List() (val []healthy_type.Handler) {
	healthList.Range(func(_, value interface{}) bool {
		val = append(val, value.(healthy_type.Handler))
		return true
	})
	return
}

func Register(name string, r healthy_type.Handler) {
	if r == nil {
		return
	}

	xerror.Assert(name == "", "[name] is null")
	xerror.Assert(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
}
