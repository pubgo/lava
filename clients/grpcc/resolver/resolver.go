package resolver

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/plugins/registry"
)

var logs = logz.New("balancer.resolver")

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
	cc      resolver.ClientConn
	r       registry.Watcher
}

func (r *baseResolver) Close() {
	r.cancel()
}

func (r *baseResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	logs.Infof("[grpc] %s ResolveNow", r.builder)
}

// 关于 grpc 命名的介绍
// https://github.com/grpc/grpc/blob/master/doc/naming.md

// BuildDirectTarget direct:///localhost:8080,localhost:8081
func BuildDirectTarget(endpoints ...string) string {
	return fmt.Sprintf("%s:///%s", DirectScheme, strings.Join(endpoints, EndpointSep))
}

// BuildDiscovTarget discov://etcd/test-service
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
		Attributes: attributes.New(),
		ServerName: name,
	}
}

// 组合服务的id和replica序列号
func getServiceUniqueId(name string, id int) string {
	return fmt.Sprintf("%s-%d", name, id)
}
