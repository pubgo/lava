package golug_broker

import (
	"sync"

	"github.com/pubgo/xerror"
)

var data sync.Map
var registerData = make(map[string]func() Broker)

func Register(name string, broker func() Broker) {
	if broker == nil {
		xerror.Next().Panic(xerror.Fmt("[broker] %s is nil", name))
	}

	if registerData[name] != nil {
		xerror.Next().Panic(xerror.Fmt("[broker] %s already exists", name))
	}

	registerData[name] = broker
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
	data.Range(func(key, value interface{}) bool { dt[key.(string)] = value.(Broker); return true })
	return dt
}
