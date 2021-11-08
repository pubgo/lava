package healthy

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/types"
)

const Name = "health"

var healthList typex.SMap

func Get(names ...string) types.Healthy {
	val, ok := healthList.Load(lavax.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(types.Healthy)
}

func List() (val []types.Healthy) {
	healthList.Range(func(_, value interface{}) bool {
		val = append(val, value.(types.Healthy))
		return true
	})
	return
}

func Register(name string, r types.Healthy) {
	if r == nil {
		return
	}

	xerror.Assert(name == "", "[name] is null")
	xerror.Assert(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
}
