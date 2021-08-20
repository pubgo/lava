package grpc_gw

import (
	"context"
	"reflect"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func Register(ctx context.Context, mux *gw.ServeMux, conn *grpc.ClientConn) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(conn == nil, "[conn] should not be nil")
	xerror.Assert(mux == nil, "[mux] should not be nil")
	xerror.Assert(ctx == nil, "[ctx] should not be nil")

	for v, val := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 3 {
			continue
		}

		if val == nil {
			continue
		}

		//if v1.In(0).String() != "context.Context" ||
		//	v1.In(1).String() != "*runtime.ServeMux" ||
		//	v1.In(2).String() != "*grpc.ClientConn" {
		//	continue
		//}

		var ff, ok = val.(func(context.Context, *gw.ServeMux, *grpc.ClientConn) error)
		if !ok {
			continue
		}

		xerror.Panic(ff(ctx, mux, conn))
	}

	return nil
}
