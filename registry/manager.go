package registry

import (
	"github.com/pubgo/golug/consts"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var registries types.SMap
var Default Registry

func List() (dt map[string]Registry) {
	xerror.Panic(registries.Map(&dt))
	return
}

func Register(r Registry) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(r == nil || r.String() == "", "[driver] is null")

	schema := r.String()
	xerror.Assert(registries.Has(schema), "registry %s already exists", schema)

	registries.Set(schema, r)
	return
}

func Get(schemas ...string) Registry {
	val, ok := registries.Load(consts.GetDefault(schemas...))
	if !ok {
		return nil
	}

	return val.(Registry)
}
