package grpccresolver

import (
	"context"

	"github.com/pubgo/funk/log"
	"google.golang.org/grpc/resolver"
)

const (
	DnsScheme       = "dns"
	K8sScheme       = "k8s"
	DirectScheme    = "direct"
	DiscoveryScheme = "discovery"
	EndpointSep     = ","
)

var logs = log.GetLogger("balancer.resolver")

var Replica = 1

type baseResolver struct {
	builder     string
	serviceName string
	cancel      context.CancelFunc
}

func (r *baseResolver) Close() {
	if r.cancel != nil {
		r.cancel()
	}
}

func (r *baseResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	logs.Info().Str("service-name", r.serviceName).Str("resolver-builder", r.builder).Msgf("grpc balancer resolve now")
}
