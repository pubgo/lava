package resolver

import (
	"fmt"
	"net"
	"strings"

	"github.com/pubgo/xerror"
	"google.golang.org/grpc/resolver"
)

var _ resolver.Builder = (*dnsResolver)(nil)

type dnsResolver struct{}

func (r *dnsResolver) Scheme() string { return "dns" }
func (r *dnsResolver) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	endpoints := strings.Split(target.URL.Host, EndpointSep)
	if len(endpoints) == 0 {
		return nil, xerror.Fmt("%s has not endpoint", target.URL.String())
	}

	// 构造resolver address, 并处理副本集
	var addrList []resolver.Address
	for i := range endpoints {
		// TODO: 考虑定时解析
		// 根据规则解析出地址
		lookups, err := net.LookupHost(endpoints[i])
		xerror.Panic(err)

		// 构造resolver address, 并处理副本集
		for j := range lookups {
			addr := lookups[j]
			for k := 0; k < Replica; k++ {
				addrList = append(addrList, newAddr(addr, fmt.Sprintf("%v", k)))
			}
		}
	}

	// 根据规则解析出地址
	if len(addrList) == 0 {
		return nil, xerror.Fmt("%v has not endpoint", target)
	}

	xerror.PanicF(cc.UpdateState(newState(addrList)), "update resolver address: %v", addrList)
	return &baseResolver{builder: "dns"}, nil
}
