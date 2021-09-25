package grpc_gw

import (
	"context"
	"reflect"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func Register(ctx context.Context, mux *gw.ServeMux, conn *grpc.ClientConn, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(conn == nil, "[conn] should not be nil")
	xerror.Assert(mux == nil, "[mux] should not be nil")
	xerror.Assert(ctx == nil, "[ctx] should not be nil")
	xerror.Assert(handler == nil, "[handler] should not be nil")

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v, val := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 3 {
			continue
		}

		if val == nil {
			continue
		}

		//func RegisterUserServiceHandler(ctx context.Context, mux *runtime.ServeMux, conn *grpc.ClientConn) error {
		//fmt.Println(hd.Implements(reflect.TypeOf(val).In(1)), v1.String(), v1.Name())
		var ff, ok = v.Interface().(func(context.Context, *gw.ServeMux, *grpc.ClientConn) error)
		if !ok || !hd.Implements(reflect.TypeOf(val).In(1)) {
			continue
		}

		xerror.Panic(ff(ctx, mux, conn))
	}

	return nil
}
