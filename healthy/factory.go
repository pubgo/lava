package healthy

import (
	"context"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"

	"github.com/pubgo/xerror"
)

const Name = "health"

type HealthCheck func(ctx context.Context) error

var healthList typex.SMap

func Get(names ...string) HealthCheck {
	val, ok := healthList.Load(consts.GetDefault(names...))
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
	defer xerror.RespExit()
	xerror.Assert(name == "" || r == nil, "[name,r] is null")
	xerror.Assert(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
	return
}
