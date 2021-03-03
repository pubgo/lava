package grpclient

import (
	"context"

	"github.com/pubgo/golug/client/grpclient/balancer/resolver"
	"github.com/pubgo/golug/registry"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc"
)

func buildTarget(service string) string {
	// 注册中心为nil走直连模式
	if registry.Default == nil {
		return resolver.BuildDirectTarget([]string{service})
	}

	return resolver.BuildDiscovTarget([]string{registry.Default.String()}, service)
}

func dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultTimeout)
	defer cancel()

	conn, err := grpc.DialContext(ctx, target, opts...)
	if err != nil {
		return nil, xerror.WrapF(err, "DialContext error, target:%s\n", target)
	}
	return conn, nil
}
