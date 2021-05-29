package grpc

import (
	"reflect"

	"github.com/pubgo/lug/registry"
)

func newRpcHandler(handler interface{}) []*registry.Endpoint {
	typ := reflect.TypeOf(handler)
	hd := reflect.ValueOf(handler)
	name := reflect.Indirect(hd).Type().Name()

	var endpoints []*registry.Endpoint

	for m := 0; m < typ.NumMethod(); m++ {
		if e := extractEndpoint(typ.Method(m)); e != nil {
			e.Name = name + "." + e.Name
			endpoints = append(endpoints, e)
		}
	}

	return endpoints
}
