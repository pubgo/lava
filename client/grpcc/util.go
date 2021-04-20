package grpcc

import (
	"github.com/pubgo/lug/client/grpcc/balancer/resolver"
	"github.com/pubgo/lug/registry"
)

func buildTarget(service string) string {
	// 注册中心为nil走直连模式
	if registry.Default() == nil {
		return resolver.BuildDirectTarget([]string{service})
	}

	return resolver.BuildDiscovTarget([]string{registry.Default().String()}, service)
}
