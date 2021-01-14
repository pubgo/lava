package grpclient

import (
	"context"

	"github.com/pubgo/golug/golug_balancer/resolver"
	registry "github.com/pubgo/golug/golug_registry"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func initOption(cfg Cfg) {

}

func buildTarget(services []string) []string {
	// 注册中心为nil
	// 走直连模式
	var targets = make([]string, 0, len(services))
	for i := range services {
		target := resolver.BuildDirectTarget([]string{services[i]})
		if registry.Get(cfg.Registry) != nil {
			target = resolver.BuildDiscovTarget([]string{registry.Default.String()}, services[i])
		}
		targets = append(targets, target)
	}

	return targets
}

func New(service string, opts ...grpc.DialOption) (grpc.ClientConnInterface, error) {
	val, ok := clients.Load(service)
	if ok {
		return val.(grpc.ClientConnInterface), nil
	}

	if err := Init([]string{service}, opts...); err != nil {
		return nil, xerror.Wrap(err)
	}

	return New(service, opts...)
}

// Init
func Init(services []string, opts ...grpc.DialOption) error {
	for i, target := range buildTarget(services) {
		service := services[i]
		_, ok := clients.Load(service)
		if ok {
			continue
		}

		conn, err := dial(target, opts...)
		if err != nil {
			return xerror.Wrap(err)
		}

		clients.Store(service, conn)
	}
	return nil
}

func dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options := getDialOption()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, target, append(options, opts...)...)
	if err != nil {
		return nil, xerror.WrapF(err, "target:%s DialContext error:%s", target, err.Error())
	}
	return conn, nil
}
