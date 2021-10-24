package grpc_gw

import (
	"context"
	"reflect"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/lava/xgen"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func Register(ctx context.Context, mux *gw.ServeMux, conn *grpc.ClientConn, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(conn == nil, "[conn] should not be nil")
	xerror.Assert(mux == nil, "[mux] should not be nil")
	xerror.Assert(ctx == nil, "[ctx] should not be nil")
	xerror.Assert(handler == nil, "[handler] should not be nil")

	for v, _ := range xgen.List() {
		if v.Type().Kind() != reflect.Func {
			continue
		}

		var ff, ok = v.Interface().(func(context.Context, *gw.ServeMux, *grpc.ClientConn) error)
		if !ok {
			continue
		}
		xerror.Panic(ff(ctx, mux, conn))
	}

	return nil
}
