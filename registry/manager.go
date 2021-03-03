package registry

import (
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var registries types.SMap
var Default Registry

func List() (dt map[string]Registry) {
	xerror.Panic(registries.Map(&dt))
	return
}

func Register(name string, r Registry) {
	xerror.Assert(name == "" || r == nil, "[name] or [r] is null")
	xerror.Assert(registries.Has(name), "registry %s already exists", name)

	registries.Set(name, r)
}

func Get(name string) Registry {
	val, ok := registries.Load(name)
	if !ok {
		return nil
	}

	return val.(Registry)
}
