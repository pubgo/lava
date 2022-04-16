package grpcc_builder

import (
	"context"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/abc"
	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/core/logging"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/plugin"
)

var logs = logging.Component(grpcc_config.Name)
var clients sync.Map

func InitClient(srv string, clientType interface{}, newClient func(cc grpc.ClientConnInterface) interface{}) {
	defer xerror.RespExit()

	logs.L().Info("grpc client init", zap.String(logkey.Service, srv))
	if val, ok := clients.LoadOrStore(srv, grpcc.NewClient(srv)); ok && val != nil {
		return
	}

	xerror.Assert(clientType == nil, "grpc clientType is nil")
	xerror.Assert(newClient == nil, "grpc newClient is nil")

	// 依赖注入
	inject.Register(clientType, func(obj inject.Object, field inject.Field) (interface{}, bool) {
		var conn, ok = clients.Load(fmt.Sprintf("%s.%s", srv, field.Name()))
		if ok {
			return newClient(conn.(grpc.ClientConnInterface)), true
		}

		logs.L().Error("grpc service not found", zap.String(logkey.Service, srv))
		return nil, false
	})
}

func CreateConn(addr string, cfg *grpcc_config.Cfg, plugins ...string) (*grpc.ClientConn, error) {
	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.DialTimeout)
	defer cancel()

	var middlewares []abc.Middleware

	// 加载全局middleware
	for _, plg := range plugins {
		middlewares = append(middlewares, plugin.Get(plg).Middleware())
	}

	conn, err := grpc.DialContext(ctx, addr, append(cfg.ToOpts(),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(middlewares)),
		grpc.WithChainStreamInterceptor(streamInterceptor(middlewares)))...)
	return conn, xerror.WrapF(err, "DialContext error, target:%s\n", addr)
}

func BuildTarget(service string, registry ...string) string {
	// 127.0.0.1,127.0.0.1,127.0.0.1;127.0.0.1
	var host = extractHostFromHostPort(service)
	var scheme = grpcc_resolver.DiscovScheme
	var reg = "mdns"
	if len(registry) > 0 {
		reg = registry[0]
	}

	if strings.Contains(service, ",") || net.ParseIP(host) != nil || host == "localhost" {
		scheme = grpcc_resolver.DirectScheme
	}

	if strings.Contains(service, "k8s://") || net.ParseIP(host) != nil || host == "localhost" {
		scheme = grpcc_resolver.DirectScheme
	}

	switch scheme {
	case grpcc_resolver.DiscovScheme:
		return grpcc_resolver.BuildDiscovTarget(service, reg)
	case grpcc_resolver.DirectScheme:
		return grpcc_resolver.BuildDirectTarget(service)
	default:
		panic("schema is unknown")
	}
}

func extractHostFromHostPort(ep string) string {
	host, _, err := net.SplitHostPort(ep)
	if err != nil {
		return ep
	}
	return host
}
