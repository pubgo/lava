package broker

import (
	"github.com/pubgo/golug/types"
	"github.com/pubgo/xerror"
)

var brokerMap = types.NewSMap()

func Register(name string, broker Broker) {
	xerror.Assert(name == "" || broker == nil, "[broker], [name] should not be null")
	xerror.Assert(brokerMap.Has(name), "[broker] %s already exists", name)

	brokerMap.Set(name, broker)
}

func Get(name string) Broker {
	val, ok := brokerMap.Load(name)
	if !ok {
		return nil
	}

	return val.(Broker)
}

func List() (dt map[string]Broker) { brokerMap.Map(&dt); return }
