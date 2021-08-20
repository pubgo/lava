package xgen

import (
	"reflect"
	"sync"
)

var data sync.Map

func Add(key reflect.Value, value interface{}) {
	data.Store(key, value)
}

func List() map[reflect.Value]interface{} {
	dt := make(map[reflect.Value]interface{})
	data.Range(func(key, value interface{}) bool {
		dt[key.(reflect.Value)] = value
		return true
	})
	return dt
}
