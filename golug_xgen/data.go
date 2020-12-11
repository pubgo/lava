package golug_xgen

import (
	"reflect"
	"sync"
)

var data sync.Map

func Add(key reflect.Value, value []GrpcRestHandler) {
	data.Store(key, value)
}

func List() map[reflect.Value][]GrpcRestHandler {
	dt := make(map[reflect.Value][]GrpcRestHandler)
	data.Range(func(key, value interface{}) bool {
		dt[key.(reflect.Value)] = value.([]GrpcRestHandler)
		return true
	})
	return dt
}
