package watcher

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/types"
)

type Factory = func(cfg typex.M) (Watcher, error)
type Handler = func(name string, r *types.WatchResp) error

var factories = make(map[string]Factory)
var callbacks = make(map[string][]Handler)

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

func Watch(name string, callback ...Handler) {
	name = KeyToDot(name)
	xerror.Assert(name == "" || len(callback) == 0, "[name, callback] should not be null")
	callbacks[name] = append(callbacks[name], callback...)
}
