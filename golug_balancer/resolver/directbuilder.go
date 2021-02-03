package resolver

import (
	"fmt"
	"strings"

	"google.golang.org/grpc/resolver"
)

type directBuilder struct{}

// directBuilder direct:///127.0.0.1,wpt.etcd:2379
func (d *directBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// 根据规则解析出地址
	endpoints := strings.FieldsFunc(target.Endpoint, func(r rune) bool { return r == EndpointSepChar })
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("%v has not endpoint", target)
	}

	// 构造resolver address, 并处理副本集
	var addrs []resolver.Address
	for i := range endpoints {
		addr := endpoints[i]
		for j := 0; j < Replica; j++ {
			addrs = append(addrs, newAddr(addr, addr))
		}
	}

	cc.UpdateState(resolver.State{Addresses: reshuffle(addrs)})
	return &baseResolver{cc: cc}, nil
}

func (d *directBuilder) Scheme() string { return DirectScheme }
