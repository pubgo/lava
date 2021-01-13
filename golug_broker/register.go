package golug_broker

import (
	"sync"

	"github.com/pubgo/xerror"
)

var brokerMap sync.Map

func Register(name string, broker Broker) {
	xerror.Assert(name == "" || broker == nil, "[broker], [name] should not be null")

	if _, ok := brokerMap.LoadOrStore(name, broker); ok {
		xerror.Assert(ok, "[broker] %s already exists", name)
	}
}

func Get(name string) Broker {
	val, ok := brokerMap.Load(name)
	if !ok {
		return nil
	}
	return val.(Broker)
}

func List() map[string]Broker {
	var dt = make(map[string]Broker)
	brokerMap.Range(func(key, value interface{}) bool { dt[key.(string)] = value.(Broker); return true })
	return dt
}
