package golug_registry

import (
	"github.com/pubgo/xerror"
	"net/url"
	"sync"
)

var registries sync.Map
var Default Registry

func Register(name string, r Registry) {
	xerror.Assert(name == "" || r == nil, "[name] or [r] is nil")

	_, ok := registries.LoadOrStore(name, r)
	xerror.Assert(ok, "registry %s is exists", name)
}

func Init(rawUrl string) {
	url1, err := url.Parse(rawUrl)
	xerror.PanicF(err, "url %s parse error", rawUrl)

	xerror.Assert(Get(scheme) == nil, "registry [%s] not exists", scheme)

	Default = Get(scheme)
	xerror.PanicF(Default.Init(opts...), "[%s] init error", scheme)
}

func Get(name string) Registry {
	val, ok := registries.Load(name)
	if !ok {
		return nil
	}

	return val.(Registry)
}

func List() map[string]Registry {
	var data = make(map[string]Registry)
	registries.Range(func(key, value interface{}) bool {
		data[key.(string)] = value.(Registry)
		return true
	})
	return data
}
