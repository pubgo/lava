package grpcs

import (
	"fmt"
	"reflect"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
	"google.golang.org/grpc/channelz/service"
	"google.golang.org/grpc/encoding"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	enc "github.com/pubgo/lug/encoding"
	"github.com/pubgo/lug/xgen"
)

func EnableHealth(srv string, s *grpc.Server) {
	healthCheck := health.NewServer()
	healthCheck.SetServingStatus(srv, grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(s, healthCheck)
}

func EnableReflection(s *grpc.Server) {
	reflection.Register(s)
}

func EnableDebug(s *grpc.Server) {
	grpc.EnableTracing = true
	service.RegisterChannelzServiceToServer(s)
}

func InitEncoding() {
	enc.Each(func(_ string, cdc enc.Codec) {
		encoding.RegisterCodec(cdc)
	})
}

// ListGRPCResources is a helper function that lists all URLs that are registered on gRPC server.
//
// This makes it easy to register all the relevant routes in your HTTP router of choice.
func ListGRPCResources(server *grpc.Server) []string {
	var ret []string
	for serviceName, serviceInfo := range server.GetServiceInfo() {
		for _, methodInfo := range serviceInfo.Methods {
			ret = append(ret, fmt.Sprintf("/%s/%s", serviceName, methodInfo.Name))
		}
	}
	return ret
}

func Register(server *grpc.Server, handler interface{}) error {
	xerror.Assert(server == nil, "[server] should not be nil")

	var v = FindHandle(handler)
	if v.IsValid() {
		_ = fx.WrapValue(v, server, handler)
		return nil
	}

	return xerror.Fmt("register [%#v] 没有找到匹配的interface", handler)
}

func FindHandle(handler interface{}) reflect.Value {
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
