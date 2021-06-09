package rest

import (
	"reflect"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
)

func register(server fiber.Router, handler interface{}) error {
	xerror.Assert(server == nil, "[server] should not be nil")

	var v = checkHandle(handler)
	if v.IsValid() {
		_ = fx.WrapValue(v, server, handler)
		return nil
	}

	return xerror.Fmt("register [%#v] 没有找到匹配的interface", handler)
}

func checkHandle(handler interface{}) reflect.Value {
	xerror.Assert(handler == nil, "[handler] should not be nil")

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 || v1.In(1).Kind() != reflect.Interface {
			continue
		}

		if !hd.Implements(v1.In(1)) || v1.In(0).String() != "fiber.Router" {
			continue
		}

		return v
	}

	return reflect.Value{}
}
