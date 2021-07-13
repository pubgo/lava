package healthy

import (
	"context"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"
	
	"github.com/pubgo/xerror"
)

type HealthCheck func(ctx context.Context) error

var healthList typex.Map

func Get(names ...string) HealthCheck {
	val, ok := healthList.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(HealthCheck)
}

func Register(name string, r HealthCheck) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "" || r == nil, "[name,r] is null")
	xerror.Assert(healthList.Has(name), "healthy [%s] already exists", name)
	healthList.Set(name, r)
	return
}
