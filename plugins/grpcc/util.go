package grpcc

import (
	"strings"

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

// serviceFromMethod returns the service
// /service.Foo/Bar => service
func serviceFromMethod(m string) string {
	if len(m) == 0 {
		return m
	}

	if m[0] != '/' {
		return m
	}

	parts := strings.Split(m, "/")
	if len(parts) < 3 {
		return m
	}

	parts = strings.Split(parts[1], ".")
	return strings.Join(parts[:len(parts)-1], ".")
}
