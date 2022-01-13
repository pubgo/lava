package tracing

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/types"
)

var factories typex.SMap

type Factory func(cfg types.CfgMap) error

func GetFactory(names ...string) Factory {
	val, ok := factories.Load(lavax.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func RegisterFactory(name string, r Factory) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "" || r == nil, "[name,tracer] is null")
	xerror.Assert(factories.Has(name), "tracer %s already exists", name)
	factories.Set(name, r)
	return
}
