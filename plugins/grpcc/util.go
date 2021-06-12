package grpcc

import (
	resolver2 "github.com/pubgo/lug/plugins/grpcc/balancer/resolver"
	"github.com/pubgo/lug/registry"
)

func buildTarget(service string) string {
	// 注册中心为nil走直连模式
	if registry.Default() == nil {
		return resolver2.BuildDirectTarget([]string{service})
	}

	return resolver2.BuildDiscovTarget([]string{registry.Default().String()}, service)
}
