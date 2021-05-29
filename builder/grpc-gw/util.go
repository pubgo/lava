package grpc_gw

import (
	"context"
	"net/http"
	"reflect"
	"time"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/lug/app"
	"github.com/pubgo/lug/xgen"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

// 开启api网关模式
func startGw(addr string) (err error) {
	gw.DefaultContextTimeout = time.Second * 2

	// 开启api网关模式
	mux := gw.NewServeMux(
		gw.WithMetadata(func(ctx context.Context, r *http.Request) metadata.MD {
			return metadata.MD(r.URL.Query())
		}),

		gw.WithMarshalerOption(gw.MIMEWildcard, &gw.HTTPBodyMarshaler{
			Marshaler: &gw.JSONPb{
				MarshalOptions: protojson.MarshalOptions{
					UseProtoNames:  true,
					UseEnumNumbers: true,
				},
				UnmarshalOptions: protojson.UnmarshalOptions{
					DiscardUnknown: true,
				},
			},
		}),
	)

	var server = &http.Server{Addr: addr, Handler: mux}

	// 注册网关api
	xerror.Panic(registerGw(app.Addr, mux, grpc.WithBlock(), grpc.WithInsecure()))

	xerror.Exit(fx.GoDelay(time.Second, func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			xlog.Error("Srv [GW] Listen Error", xlog.Any("err", err))
		}

		xlog.Info("Srv [GW] Closed OK")
	}))

	xlog.Infof("Srv [GW] Listening on http://localhost%s", app.Addr)

	//g.BeforeStop(func() {
	//	if err := server.Shutdown(context.Background()); err != nil {
	//		xlog.Error("Srv [GW] Shutdown Error", xlog.Any("err", err))
	//	}
	//})

	return nil
}

func registerGw(srv string, mux *gw.ServeMux, opts ...grpc.DialOption) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(mux == nil, "[mux] should not be nil")
	xerror.Assert(srv == "", "[srv] should not be null")

	var params = []interface{}{context.Background(), mux, srv}
	for i := range opts {
		params = append(params, opts[i])
	}

	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 3 {
			continue
		}

		if v.Type().In(1).String() != "runtime.ServeMux" {
			continue
		}

		_ = fx.WrapValue(v, params...)
	}
	return
}