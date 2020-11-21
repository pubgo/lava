package grpc_entry

import (
	"github.com/pubgo/golug/golug_data"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"reflect"
)

func register(server *grpc.Server, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	if handler == nil {
		return xerror.New("[handler] should not be nil")
	}

	if server == nil {
		return xerror.New("[server] should not be nil")
	}

	var vRegister reflect.Value
	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for _, v := range golug_data.List() {
		v1 := reflect.TypeOf(v)
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 {
			continue
		}

		if hd.Implements(v1.In(1)) {
			vRegister = reflect.ValueOf(v)
			break
		}
	}

	if !vRegister.IsValid() || vRegister.IsNil() {
		return xerror.Fmt("[%#v, %#v] 没有找到匹配的interface", handler, vRegister.Interface())
	}

	vRegister.Call([]reflect.Value{reflect.ValueOf(server), reflect.ValueOf(handler)})
	return
}
