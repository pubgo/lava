package golug_registry

import (
	"github.com/pubgo/golug/golug_types"
	"github.com/pubgo/xerror"
)

var registries = golug_types.NewSyncMap()
var Default Registry

func Register(name string, r Registry) {
	xerror.Assert(name == "" || r == nil, "[name] or [r] is nil")
	xerror.Assert(registries.Has(name), "registry %s is exists", name)

	registries.Set(name, r)
}

func Get(name string) Registry {
	val, ok := registries.Load(name)
	if !ok {
		return nil
	}

	return val.(Registry)
}

func List() (dt map[string]Registry) { registries.Map(&dt); return }
