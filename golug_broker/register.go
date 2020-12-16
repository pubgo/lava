package golug_broker

import (
	"sync"

	"github.com/pubgo/xerror"
)

var data sync.Map

func Register(name string, broker Broker) {
	if broker == nil {
		xerror.Next().Panic(xerror.Fmt("[broker] %s is nil", name))
	}

	if _, ok := data.LoadOrStore(name, broker); ok {
		xerror.Next().Panic(xerror.Fmt("[broker] %s already exists", name))
	}
}

func Get(name string) Broker {
	val, ok := data.Load(name)
	if ok {
		return val.(Broker)
	}
	return nil
}

func List() map[string]Broker {
	var dt = make(map[string]Broker)
	data.Range(func(key, value interface{}) bool {
		dt[key.(string)] = value.(Broker)
		return true
	})
	return dt
}
