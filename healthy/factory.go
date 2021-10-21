package healthy

import (
	"context"
	"github.com/pubgo/lava/pkg/lavax"

	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/typex"
)

const Name = "health"

type HealthCheck func(ctx context.Context) error

var healthList typex.SMap

func Get(names ...string) HealthCheck {
	val, ok := healthList.Load(lavax.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(HealthCheck)
}

func List() (val []HealthCheck) {
	healthList.Range(func(_, value interface{}) bool {
		val = append(val, value.(HealthCheck))
		return true
	})
	return
}

func Register(name string, r HealthCheck) {
	if r == nil {
		return
	}

	xerror.Assert(name == "", "[name] is null")
	xerror.Assert(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
}
