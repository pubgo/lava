package resolver

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/pubgo/lug/registry"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

const (
	DirectScheme    = "direct"
	DiscovScheme    = "discov"
	EndpointSepChar = ','
)

var (
	EndpointSep = fmt.Sprintf("%c", EndpointSepChar)
	Replica     = 1
)

func init() {
	resolver.Register(&directBuilder{})
	resolver.Register(&discovBuilder{})
}

type baseResolver struct {
	cancel context.CancelFunc
	cc     resolver.ClientConn
	r      registry.Watcher
}

func (r *baseResolver) Close() {
	defer r.cancel()
	if r.r != nil {
		r.r.Stop()
	}
}

func (r *baseResolver) ResolveNow(_ resolver.ResolveNowOptions) {
	logs.Info("[grpc] ResolveNow")
}

// 关于 grpc 命名的介绍
// https://github.com/grpc/grpc/blob/master/doc/naming.md

func BuildDirectTarget(endpoints ...string) string {
	return fmt.Sprintf("%s:///%s", DirectScheme, strings.Join(endpoints, EndpointSep))
}

func BuildDiscovTarget(service string, endpoints ...string) string {
	return fmt.Sprintf("%s://%s/%s", DiscovScheme, strings.Join(endpoints, EndpointSep), service)
}

//对targets打散
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
