package golug_data

import "sync"

var data sync.Map

func Add(key, value interface{}) {
	data.Store(key, value)
}

func Delete(key interface{}) {
	data.Delete(key)
}

func Update(key interface{}, value interface{}) {
	data.Store(key, value)
}

func Get(key interface{}) (interface{}, bool) {
	return data.Load(key)
}

func List() map[interface{}]interface{} {
	dt := make(map[interface{}]interface{})
	data.Range(func(key, value interface{}) bool {
		dt[key] = value
		return true
	})
	return dt
}
