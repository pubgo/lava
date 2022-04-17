package grpcc_resolver

import (
	"net"
	"strings"

	"google.golang.org/grpc/resolver"
)

func newState(addrList []resolver.Address) resolver.State {
	return resolver.State{Addresses: reshuffle(addrList)}
}

func BuildTarget(service string, registry ...string) string {
	// 127.0.0.1,127.0.0.1,127.0.0.1;127.0.0.1
	var host = extractHostFromHostPort(service)
	var scheme = DiscovScheme
	var reg = "mdns"
	if len(registry) > 0 {
		reg = registry[0]
	}

	if strings.Contains(service, ",") || net.ParseIP(host) != nil || host == "localhost" {
		scheme = DirectScheme
	}

	if strings.Contains(service, "k8s://") || net.ParseIP(host) != nil || host == "localhost" {
		scheme = K8sScheme
	}

	switch scheme {
	case DiscovScheme:
		return BuildDiscovTarget(service, reg)
	case DirectScheme:
		return BuildDirectTarget(service)
	default:
		panic("schema is unknown")
	}
}
