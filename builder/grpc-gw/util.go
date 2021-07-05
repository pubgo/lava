package grpc_gw

import (
	"github.com/pubgo/lug/xgen"

	"context"
	"reflect"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func Register(ctx context.Context, mux *gw.ServeMux, conn *grpc.ClientConn) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(conn == nil, "[conn] should not be nil")
	xerror.Assert(mux == nil, "[mux] should not be nil")
	xerror.Assert(ctx == nil, "[ctx] should not be nil")

	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 3 {
			continue
		}

		if v1.In(0).String() != "context.Context" ||
			v1.In(1).String() != "runtime.ServeMux" ||
			v1.In(2).String() != "grpc.ClientConn" {
			continue
		}

		fx.Wrap(v)(ctx, mux, conn)(func(err2 error) { xerror.Panic(err2) })
	}

	return nil
}
