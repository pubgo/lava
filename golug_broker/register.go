package golug_broker

import "sync"

var data sync.Map

func Register(name string, broker Broker) {
	data.Store(name, broker)
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
