package registry

import (
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var registries = types.NewSyncMap()
var Default Registry

func List() (dt map[string]Registry) { registries.Map(&dt); return }
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
