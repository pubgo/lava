package restEntry

import (
	"reflect"
	"strings"

	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/xgen"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/x/byteutil"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
)

func register(server fiber.Router, handler interface{}) (err error) {
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
			v1.In(0).String() != "fiber.Router" ||
			v1.In(1).Kind() != reflect.Interface ||
			!hd.Implements(v1.In(1)) {
			continue
		}

		return v
	}

	return reflect.Value{}
}

func convertHeader(request interface{ VisitAll(func(key, value []byte)) }) types.Header {
	var h = types.HeaderGet()
	request.VisitAll(func(key, value []byte) {
		h.Add(byteutil.ToStr(key), byteutil.ToStr(value))
	})
	return h
}

func getPort(addr string) string {
	var addrList = strings.Split(addr, ":")
	return addrList[len(addrList)-1]
}
