package grpcc_resolver

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/pubgo/funk/log"
	"google.golang.org/grpc/attributes"
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
	builder string
	name    string
	cancel  context.CancelFunc
}

func (r *baseResolver) Close() {
	if r.cancel == nil {
		return
	}
	r.cancel()
}

func (r *baseResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	logs.Info().Msgf("[grpc] %s ResolveNow", r.builder)
}

// gRPC名称解析
// 	https://github.com/grpc/grpc/blob/master/doc/naming.md
// 	dns:[//authority/]host[:port]

// BuildDirectTarget direct://localhost:8080,localhost:8081
func BuildDirectTarget(name string, endpoints ...string) string {
	return fmt.Sprintf("%s://%s?name=%s", DirectScheme, strings.Join(endpoints, EndpointSep), name)
}

// BuildDiscoveryTarget discovery://test-service:8080
func BuildDiscoveryTarget(service string) string {
	return fmt.Sprintf("%s://%s", DiscoveryScheme, service)
}

// reshuffle 打散targets
func reshuffle(targets []resolver.Address) []resolver.Address {
	rand.Shuffle(len(targets), func(i, j int) { targets[i], targets[j] = targets[j], targets[i] })
	return targets
}

// 创建新的Address
func newAddr(addr string, name string) resolver.Address {
	return resolver.Address{
		Addr:       addr,
		Attributes: attributes.New(addr, name),
		ServerName: name,
	}
}

// 组合服务的id和replica序列号
func getServiceUniqueId(name string, id int) string {
	return fmt.Sprintf("%s-%d", name, id)
}
