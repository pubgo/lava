package golug_grpc

import (
	"reflect"

	registry "github.com/pubgo/golug/registry"
)

func newRpcHandler(handler interface{}) []*registry.Endpoint {
	typ := reflect.TypeOf(handler)
	hdlr := reflect.ValueOf(handler)
	name := reflect.Indirect(hdlr).Type().Name()

	var endpoints []*registry.Endpoint

	for m := 0; m < typ.NumMethod(); m++ {
		if e := extractEndpoint(typ.Method(m)); e != nil {
			e.Name = name + "." + e.Name
			endpoints = append(endpoints, e)
		}
	}

	return endpoints
}
