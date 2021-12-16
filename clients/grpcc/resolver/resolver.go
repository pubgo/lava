package resolver

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"

	"github.com/pubgo/lava/logz"
)

var logs = logz.Component("balancer.resolver")

const (
	DirectScheme = "direct"
	DiscovScheme = "discov"
	EndpointSep  = ","
)

var (
	Replica = 1
)

type baseResolver struct {
	builder string
	cancel  context.CancelFunc
}

func (r *baseResolver) Close() {
	if r.cancel == nil {
		return
	}
	r.cancel()
}

func (r *baseResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	logs.Infof("[grpc] %s ResolveNow", r.builder)
}

// gRPC名称解析
// 	https://github.com/grpc/grpc/blob/master/doc/naming.md
// 	dns:[//authority/]host[:port]

// BuildDirectTarget direct:///localhost:8080,localhost:8081
func BuildDirectTarget(endpoints ...string) string {
	return fmt.Sprintf("%s:///%s", DirectScheme, strings.Join(endpoints, EndpointSep))
}

// BuildDiscovTarget discov:///test-service
func BuildDiscovTarget(service string, endpoints ...string) string {
	return fmt.Sprintf("%s://%s/%s", DiscovScheme, strings.Join(endpoints, EndpointSep), service)
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
