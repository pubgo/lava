package client

import (
	"context"
	"fmt"

	"github.com/pubgo/golug/golug_balancer/p2c"
	"github.com/pubgo/golug/golug_balancer/resolver"
	registry "github.com/pubgo/golug/golug_registry"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

func init() {
	resolver.RegisterResolver()
	p2c.RegisterBalancer()
}

func buildTarget(services []string) []string {
	if registry.Default == nil {
		return []string{resolver.BuildDirectTarget(services)}
	}

	var targets = make([]string, len(services))
	for i := range services {
		targets = append(targets, resolver.BuildDiscovTarget([]string{registry.Default.String()}, services[i]))
	}

	return targets
}

// InitClient
func InitClient(services []string, opts ...grpc.DialOption) error {
	for i, target := range buildTarget(services) {
		conn, err := dial(target, opts...)
		if err != nil {
			return err
		}
		_, ok := clients.LoadOrStore(services[i], conn)
		if ok {
			return fmt.Errorf("%s had existed", services[i])
		}
	}
	return nil
}

func InitClientWithName(name string, service []string, opts ...grpc.DialOption) error {
	target := buildTarget(service)[0]
	conn, err := dial(target, opts...)
	if err != nil {
		return err
	}
	_, ok := clients.LoadOrStore(name, conn)
	if ok {
		return fmt.Errorf("%s had existed", name)
	}
	return nil
}

func dial(target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	options := getDialOption()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	conn, err := grpc.DialContext(ctx, target, append(options, opts...)...)
	if err != nil {
		return nil, errors.Wrapf(err, "target:%s DialContext error:%s", target, err.Error())
	}
	return conn, nil
}
