package broker

import (
	"github.com/pubgo/lug/consts"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

var factories types.SMap
var brokers types.SMap

func List() (dt map[string]Factory) { xerror.Panic(factories.Map(&dt)); return }
func Register(name string, broker Factory) {
	xerror.Assert(name == "" || broker == nil, "[broker,name] should not be null")
	xerror.Assert(factories.Has(name), "[broker] %s already exists, refer: %s", name, stack.Func(factories.Get(name)))
	factories.Set(name, broker)
}

func Get(names ...string) Broker {
	val, ok := brokers.Load(consts.GetDefault(names...))
	if !ok {
		return nil
	}
	return val.(Broker)
}
