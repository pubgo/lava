package resolver

import (
	"fmt"
	"math/rand"
	"strings"

	"github.com/pubgo/golug/registry"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc/attributes"
	"google.golang.org/grpc/resolver"
)

const (
	//direct connect to service, it is can be used in k8s or other systems without service discovery
	DirectScheme    = "direct"
	DiscovScheme    = "discov"
	EndpointSepChar = ','
)

var (
	EndpointSep = fmt.Sprintf("%c", EndpointSepChar)
	Replica     = 2
)

func init() {
	resolver.Register(&directBuilder{})
	resolver.Register(&discovBuilder{})
}

type baseResolver struct {
	cc resolver.ClientConn
	r  registry.Watcher
}

func (r *baseResolver) Close() {
	if r.r != nil {
		r.r.Stop()
	}
}

func (r *baseResolver) ResolveNow(options resolver.ResolveNowOptions) {
	xlog.Info("[grpc] ResolveNow")
}

// 关于 grpc 命名的介绍
// https://github.com/grpc/grpc/blob/master/doc/naming.md

func BuildDirectTarget(endpoints []string) string {
	return fmt.Sprintf("%s:///%s", DirectScheme, strings.Join(endpoints, EndpointSep))
}

func BuildDiscovTarget(endpoints []string, key string) string {
	return fmt.Sprintf("%s://%s/%s", DiscovScheme, strings.Join(endpoints, EndpointSep), key)
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
