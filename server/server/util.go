package server

import (
	"context"
	"net"
	"reflect"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/xgen"
)

// registerGw 找到<func(ctx context.Context, mux *runtime.ServeMux, conn grpc.ClientConnInterface) error>
func registerGw(ctx context.Context, mux *gw.ServeMux, conn grpc.ClientConnInterface, handler interface{}) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(conn == nil, "[conn] should not be nil")
	xerror.Assert(mux == nil, "[mux] should not be nil")
	xerror.Assert(ctx == nil, "[ctx] should not be nil")
	xerror.Assert(handler == nil, "[handler] should not be nil")

	for v := range xgen.List() {
		if v.Type().Kind() != reflect.Func {
			continue
		}

		var ff, ok = v.Interface().(func(context.Context, *gw.ServeMux, grpc.ClientConnInterface) error)
		if !ok {
			continue
		}
		xerror.Panic(ff(ctx, mux, conn))
	}

	return nil
}

// registerGrpc 找到<Register${Srv}Server(s grpc.ServiceRegistrar, srv ${Srv}Server)>
func registerGrpc(server grpc.ServiceRegistrar, handler interface{}) error {
	xerror.Assert(server == nil, "[server] should not be nil")

	var v = findGrpcHandle(handler)
	if v.IsValid() {
		_ = fx.WrapValue(v, server, handler)
		return nil
	}

	return xerror.Fmt("register [%#v] 没有找到匹配的interface", handler)
}

func findGrpcHandle(handler interface{}) reflect.Value {
	xerror.Assert(handler == nil, "[handler] should not be nil")

	hd := reflect.New(reflect.Indirect(reflect.ValueOf(handler)).Type()).Type()
	for v := range xgen.List() {
		v1 := v.Type()
		if v1.Kind() != reflect.Func || v1.NumIn() < 2 || v1.In(1).Kind() != reflect.Interface {
			continue
		}

		if !hd.Implements(v1.In(1)) || v1.In(0).String() != "grpc.ServiceRegistrar" {
			continue
		}

		return v
	}

	return reflect.Value{}
}

// getPeerName 获取对端应用名称
func getPeerName(md metadata.MD) string {
	return types.HeaderGet(md, "app")
}

// getPeerIP 获取对端ip
func getPeerIP(md metadata.MD, ctx context.Context) string {
	clientIP := types.HeaderGet(md, "client-ip")
	if clientIP != "" {
		return clientIP
	}

	// 从grpc里取对端ip
	pr, ok2 := peer.FromContext(ctx)
	if !ok2 {
		return ""
	}

	if pr.Addr == net.Addr(nil) {
		return ""
	}

	addSlice := strings.Split(pr.Addr.String(), ":")
	if len(addSlice) > 1 {
		return addSlice[0]
	}
	return ""
}

func ignoreMuxError(err error) bool {
	if err == nil {
		return true
	}
	return strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "mux: server closed")
}
