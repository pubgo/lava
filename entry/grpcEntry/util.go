package grpcEntry

import (
	"context"
	"fmt"
	"net"
	"reflect"
	"strings"

	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"

	gw "github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/xgen"
)

func extractValue(v reflect.Type) *registry.Value {
	defer xerror.RespExit("extractValue")

	if v == nil {
		return nil
	}

	arg := &registry.Value{
		Name: v.Name(),
		Type: v.Name(),
	}

	switch v.Kind() {
	case reflect.Ptr:
		v = v.Elem()
		arg.Name = v.Name()
		arg.Type = v.Name()

		switch v.Kind() {
		case reflect.Struct:
			for i := 0; i < v.NumField(); i++ {
				f := v.Field(i)
				val := extractValue(f.Type)
				if val == nil {
					continue
				}

				// if we can find a json tag use it
				if tags := f.Tag.Get("json"); len(tags) > 0 {
					parts := strings.Split(tags, ",")
					if parts[0] == "-" || parts[0] == "omitempty" {
						continue
					}
					val.Name = parts[0]
				}

				// if there's no name default it
				if len(val.Name) == 0 {
					val.Name = v.Field(i).Name
				}

				arg.Values = append(arg.Values, val)
			}
		case reflect.Slice:
			p := v.Elem()
			if p.Kind() == reflect.Ptr {
				p = p.Elem()
			}
			arg.Type = "[]" + p.Name()
			val := extractValue(v.Elem())
			if val != nil {
				arg.Values = append(arg.Values, val)
			}
		}
	case reflect.Interface:
		if m, ok := v.MethodByName("SendAndClose"); ok {
			arg.Values = append(arg.Values, extractValue(m.Type.In(0)))
		}

		if m, ok := v.MethodByName("Send"); ok {
			arg.Values = append(arg.Values, extractValue(m.Type.In(0)))
		}

		if m, ok := v.MethodByName("Recv"); ok {
			arg.Values = append(arg.Values, extractValue(m.Type.Out(0)))
		}
	}

	return arg
}

func extractEndpoint(method reflect.Method) *registry.Endpoint {
	defer xerror.RespExit("extractEndpoint")

	if method.PkgPath != "" {
		return nil
	}

	var rspType, reqType reflect.Type
	mt := method.Type

	var reqStream bool
	var respStream bool
	switch mt.NumOut() {
	case 1:
		switch mt.NumIn() {
		case 2:
			reqStream = true
			reqType = mt.In(1)
			rspType = mt.In(1)
			if _, ok := reqType.MethodByName("SendAndClose"); !ok {
				respStream = true
			}
		case 3:
			reqType = mt.In(1)
			rspType = mt.In(2)
			respStream = true
		}
	case 2:
		reqType = mt.In(2)
		rspType = mt.Out(0)
	}

	xerror.Assert(rspType == nil, "[rspType] is nil")

	request := extractValue(reqType)
	response := extractValue(rspType)

	return &registry.Endpoint{
		Name:     method.Name,
		Request:  request,
		Response: response,
		Metadata: map[string]string{
			"req_stream":  fmt.Sprintf("%v", reqStream),
			"resp_stream": fmt.Sprintf("%v", respStream),
		},
	}
}

func newRpcHandler(handler interface{}) []*registry.Endpoint {
	typ := reflect.TypeOf(handler)
	hd := reflect.ValueOf(handler)
	name := reflect.Indirect(hd).Type().Name()

	var endpoints []*registry.Endpoint

	for m := 0; m < typ.NumMethod(); m++ {
		if e := extractEndpoint(typ.Method(m)); e != nil {
			e.Name = name + "." + e.Name
			endpoints = append(endpoints, e)
		}
	}

	return endpoints
}

func registerGw(ctx context.Context, mux *gw.ServeMux, conn *grpc.ClientConn, handler interface{}) (err error) {
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

func registerGrpc(server *grpc.Server, handler interface{}) error {
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
