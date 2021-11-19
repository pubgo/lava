package ginEntry

import (
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/xgen"
)

func register(server gin.IRouter, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

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

		if v1.Kind() != reflect.Func ||
			v1.NumIn() < 2 ||
			v1.In(0).String() != "gin.IRouter" ||
			v1.In(1).Kind() != reflect.Interface ||
			!hd.Implements(v1.In(1)) {
			continue
		}

		return v
	}

	return reflect.Value{}
}
