package cache

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/pkg/typex"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

var factories typex.Map

type Factory func(cfg map[string]interface{}) (IStore, error)

func Register(name string, store Factory) {
	xerror.Assert(name == "" || store == nil, "[store,name] should not be null")
	xerror.Assert(factories.Has(name), "[store] %s already exists, refer: %s", name, stack.Func(factories.Get(name)))
	factories.Set(name, store)
}

func GetStore(names ...string) IStore {
	val, ok := factories.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}

	return val.(IStore)
}
