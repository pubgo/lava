package grpcc_resolver

import (
	"fmt"
	"strings"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc/resolver"
)

var _ resolver.Builder = (*directBuilder)(nil)

// directBuilder creates a directBuilder which is used to factory direct resolvers.
// example:
//   direct://127.0.0.1:9000,127.0.0.2:9000
type directBuilder struct{}

func (d *directBuilder) Scheme() string { return DirectScheme }

// Build [direct://127.0.0.1,etcd:2379]
func (d *directBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, err error) {
	defer xerror.RecoverErr(&err)

	// 根据规则解析出地址
	endpoints := strings.Split(target.URL.Host, EndpointSep)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("%s has not endpoint", target.URL.String())
	}

	// 构造resolver address, 并处理副本集
	var addrList []resolver.Address
	for i := range endpoints {
		addr := endpoints[i]
		for j := 0; j < Replica; j++ {
			addrList = append(addrList, newAddr(addr, addr))
		}
	}
	xerror.PanicF(cc.UpdateState(newState(addrList)), "update resolver address: %v", addrList)
	return &baseResolver{builder: DirectScheme}, nil
}
