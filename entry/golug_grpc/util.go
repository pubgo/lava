package golug_grpc

import (
	"context"
	"os"
	"reflect"

	gr "github.com/grpc-ecosystem/grpc-gateway/runtime"
	"github.com/pubgo/golug/xgen"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func registerGw(srv string, mux *gr.ServeMux, opts ...grpc.DialOption) (err error) {
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

func checkHandle(handler interface{}) error {
	return xutil.Try(func() {
		xerror.Assert(handler == nil, "[handler] should not be nil")

		hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
		for v := range xgen.List() {
			v1 := v.Type()
			if v1.Kind() != reflect.Func || v1.NumIn() < 2 {
				continue
			}

			if !hd.Implements(v1.In(1)) || v1.In(0).String() != "*grpc.Server" {
				continue
			}

			return
		}

		xerror.Assert(true, "[%#v] 没有找到匹配的interface", handler)
	})
}

func register(server *grpc.Server, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(server == nil, "[server] should not be nil")
	xerror.Assert(handler == nil, "[handler] should not be nil")

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 {
			continue
		}

		if !hd.Implements(v1.In(1)) || v1.In(0).String() != "*grpc.Server" {
			continue
		}

		_ = fx.WrapValue(v, server, handler)
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
