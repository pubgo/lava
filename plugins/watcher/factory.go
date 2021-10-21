package watcher

import (
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/types"
)

type Factory func(cfg typex.M) (Watcher, error)

var factories = make(map[string]Factory)
var callbacks typex.Map

func Register(name string, w Factory) {
	xerror.Assert(name == "" || w == nil, "watcher [name,w] should not be null")
	xerror.Assert(factories[name] != nil, "watcher [%s] already exists", name)
	factories[name] = w
}

func Get(names ...string) Factory {
	val, ok := factories[lavax.GetDefault(names...)]
	if !ok {
		return nil
	}

	return val
}

func Watch(name string, callback func(name string, r *types.WatchResp) error) {
	name = KeyToDot(name)
	xerror.Assert(name == "" || callback == nil, "[name, callback] should not be null")
	xerror.Assert(callbacks.Has(name), "callback [%s] already exists", name)
	callbacks.Set(name, callback)
}
