package grpcc

import (
	"github.com/pubgo/lug/plugins/grpcc/balancer/resolver"

	"github.com/pubgo/lug/registry"
)

func buildTarget(service string) string {
	// 注册中心为nil走直连模式
	if registry.Default() == nil {
		return resolver.BuildDirectTarget(service)
	}

	return resolver.BuildDiscovTarget(service, registry.Default().String())
}
