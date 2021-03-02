package golug_grpc

import (
	"os"
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func registerGw(srv string, g fiber.Router, opts ...grpc.DialOption) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(g == nil, "[g] should not be nil")
	xerror.Assert(srv == "", "[srv] should not be null")

	var paramsIn = []reflect.Value{reflect.ValueOf(srv), reflect.ValueOf(g)}
	for i := range opts {
		paramsIn = append(paramsIn, reflect.ValueOf(opts[i]))
	}

	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 3 {
			continue
		}

		// TODO check
		if v.Type().In(1).String() != "fiber.Router" {
			continue
		}

		v.Call(paramsIn)
	}
	return nil
}

func checkHandle(handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(handler == nil, "[handler] should not be nil")

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 {
			continue
		}

		if !hd.Implements(v1.In(1)) {
			continue
		}

		if v1.In(0).String() != "*grpc.Server" {
			continue
		}

		return nil
	}

	return xerror.Fmt("[%#v] 没有找到匹配的interface", handler)
}

func register(server *grpc.Server, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(handler == nil, "[handler] should not be nil")
	xerror.Assert(server == nil, "[server] should not be nil")

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 {
			continue
		}

		if !hd.Implements(v1.In(1)) {
			continue
		}

		v.Call([]reflect.Value{reflect.ValueOf(server), reflect.ValueOf(handler)})
		return nil
	}

	return xerror.Fmt("[%#v] 没有找到匹配的interface", handler)
}

func getHostname() string {
	if name, err := os.Hostname(); err != nil {
		return "unknown"
	} else {
		return name
	}
}
