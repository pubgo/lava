package watcher

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/xerror"
)

type Factory func(cfg typex.M) (Watcher, error)

var factories = make(map[string]Factory)

func Register(name string, w Factory) {
	xerror.Assert(name == "" || w == nil, "watcher [name,w] should not be null")
	xerror.Assert(factories[name] != nil, "watcher [name] %s already exists", name)
	factories[name] = w
}

func Get(names ...string) Factory {
	val, ok := factories[consts.GetDefault(names...)]
	if !ok {
		return nil
	}

	return val
}
