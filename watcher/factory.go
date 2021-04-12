package watcher

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/x/typex"
	"github.com/pubgo/xerror"
)

type Factory func(cfg typex.M) (Watcher, error)

var factories types.SMap

func Register(name string, w Factory) {
	xerror.Assert(name == "" || w == nil, "watcher [name,w] should not be null")
	xerror.Assert(factories.Has(name), "watcher [name] %s already exists", name)
	factories.Set(name, w)
}

func Get(names ...string) Factory {
	val, ok := factories.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(Factory)
}
