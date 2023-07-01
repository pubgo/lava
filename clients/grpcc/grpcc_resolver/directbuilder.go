package grpcc_resolver

import (
	"fmt"
	"strings"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"google.golang.org/grpc/resolver"
)

func NewDirectBuilder() resolver.Builder {
	return &directBuilder{
		log: logs.WithName(DirectScheme),
	}
}

var _ resolver.Builder = (*directBuilder)(nil)

// directBuilder creates a directBuilder which is used to factory direct resolvers.
// example:
//
//	direct://127.0.0.1:9000,127.0.0.2:9000?name=test-srv
type directBuilder struct {
	log log.Logger
}

func (d *directBuilder) Scheme() string { return DirectScheme }

func (d *directBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, err error) {
	defer recovery.Err(&err)

	// 根据规则解析出地址
	endpoints := strings.Split(target.URL.Host, EndpointSep)
	if len(endpoints) == 0 {
		return nil, fmt.Errorf("%s has not endpoint", target.URL.String())
	}

	var name = target.URL.Query().Get("name")

	// 构造resolver address, 并处理副本集
	var addrList []resolver.Address
	for i := range endpoints {
		addr := endpoints[i]
		for j := 0; j < Replica; j++ {
			addrList = append(addrList, newAddr(addr, name))
		}
	}

	assert.MustF(cc.UpdateState(newState(addrList)), "failed to update resolver address, address=%v", addrList)
	return &baseResolver{builder: DirectScheme, serviceName: name}, nil
}
