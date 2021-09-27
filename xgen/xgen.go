package xgen

import (
	"reflect"
	"sync"
)

type GrpcRestHandler struct {
	Input        interface{} `json:"input"`
	Output       interface{} `json:"output"`
	Service      string      `json:"service"`
	Method       string      `json:"method"`
	Name         string      `json:"name"`
	Path         string      `json:"path"`
	ClientStream bool        `json:"client_stream"`
	ServerStream bool        `json:"server_stream"`
	DefaultUrl   bool        `json:"default_url"`
}

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
