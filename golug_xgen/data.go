package golug_xgen

import (
	"reflect"
	"sync"

	"github.com/pubgo/golug/golug_entry"
)

var data sync.Map

func Add(key reflect.Value, value ...golug_entry.GrpcRestHandler) {
	data.Store(key, value)
}

func List() map[reflect.Value][]golug_entry.GrpcRestHandler {
	dt := make(map[reflect.Value][]golug_entry.GrpcRestHandler)
	data.Range(func(key, value interface{}) bool {
		dt[key.(reflect.Value)] = value.([]golug_entry.GrpcRestHandler)
		return true
	})
	return dt
}
