package grpcc_builder

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/middleware"
	"net"
	"strings"
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/pubgo/lava/clients/grpcc"
	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/clients/grpcc/grpcc_resolver"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/plugin"
)

var logs = logging.Component(grpcc_config.Name)
var clients sync.Map

func InitClient(srv string, clientType interface{}, newClient func(cc grpc.ClientConnInterface) interface{}) {
	defer xerror.RespExit()

	xerror.Assert(clientType == nil, "grpc clientType is nil")
	xerror.Assert(newClient == nil, "grpc newClient is nil")

	logs.L().Info("grpc client init", zap.String(logkey.Service, srv))
	var cli = grpcc.NewClient(srv, grpcc.WithDial(CreateConn))
	if val, ok := clients.LoadOrStore(srv, cli); ok && val != nil {
		return
	}

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

func CreateConn(srv string, cfg grpcc_config.Cfg) (grpc.ClientConnInterface, error) {
	// 创建grpc client
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Client.DialTimeout)
	defer cancel()

	var middlewares []middleware.Middleware

	// 加载全局middleware
	for _, plg := range cfg.Plugins {
		xerror.Assert(plugin.Get(plg) == nil, "plugin(%s) is nil", plg)
		if plugin.Get(plg).Middleware() == nil {
			continue
		}

		middlewares = append(middlewares, plugin.Get(plg).Middleware())
	}

	addr := BuildTarget(srv, cfg)

	conn, err := grpc.DialContext(ctx, addr, append(cfg.Client.ToOpts(),
		grpc.WithChainUnaryInterceptor(unaryInterceptor(middlewares)),
		grpc.WithChainStreamInterceptor(streamInterceptor(middlewares)))...)
	return conn, xerror.WrapF(err, "DialContext error, target:%s\n", addr)
}

func BuildTarget(service string, cfg grpcc_config.Cfg) string {
	var addr = service
	if cfg.Addr != "" {
		addr = cfg.Addr
	}

	if cfg.Registry == "" {
		cfg.Registry = "mdns"
	}

	// 127.0.0.1,127.0.0.1,127.0.0.1;127.0.0.1
	var host = extractHostFromHostPort(addr)
	var scheme = grpcc_resolver.DiscovScheme

	if strings.Contains(service, ",") || net.ParseIP(host) != nil || host == "localhost" {
		scheme = grpcc_resolver.DirectScheme
	}

	if strings.HasPrefix(service, "k8s://") {
		scheme = grpcc_resolver.K8sScheme
	}

	switch scheme {
	case grpcc_resolver.DiscovScheme:
		return grpcc_resolver.BuildDiscovTarget(service, cfg.Registry)
	case grpcc_resolver.DirectScheme:
		return grpcc_resolver.BuildDirectTarget(service)
	case grpcc_resolver.K8sScheme:
		return fmt.Sprintf("dns:///%s", service)
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
