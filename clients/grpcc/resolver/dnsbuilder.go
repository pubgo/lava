package resolver

import (
	"fmt"
	"net"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc/resolver"
)

var _ resolver.Builder = (*dnsResolver)(nil)

type dnsResolver struct{}

func (r *dnsResolver) Scheme() string { return "dns" }
func (r *dnsResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// TODO: 考虑定时解析
	lookups, err := net.LookupHost(target.Endpoint)
	xerror.Panic(err)

	// 根据规则解析出地址
	if len(lookups) == 0 {
		return nil, xerror.Fmt("%v has not endpoint", target)
	}

	// 构造resolver address, 并处理副本集
	var addrList []resolver.Address
	for i := range lookups {
		addr := lookups[i]
		for j := 0; j < Replica; j++ {
			addrList = append(addrList, newAddr(addr, fmt.Sprintf("%v", j)))
		}
	}

	xerror.PanicF(cc.UpdateState(newState(addrList)), "update resolver address: %v", addrList)
	return &baseResolver{builder: "dns"}, nil
}
