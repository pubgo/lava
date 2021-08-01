package tracing

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"
)

var factories typex.SMap

type Factory func(cfg map[string]interface{}) error

func Get(names ...string) Factory {
	val, ok := factories.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}

func Register(name string, r Factory) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(name == "" || r == nil, "[name,tracer] is null")
	xerror.Assert(factories.Has(name), "tracer %s already exists", name)
	factories.Set(name, r)
	return
}
