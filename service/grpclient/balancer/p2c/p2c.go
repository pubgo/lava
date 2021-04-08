package p2c

import (
	"context"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type p2cBalancer struct{}

func (p2c *p2cBalancer) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker { //nolint:staticcheck
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	// 创建一个新的负载均衡器
	npa := NewP2cAgl()
	for _, subConn := range readySCs {
		npa.Add(subConn)
	}

	return &p2cPicker{pickerAgl: npa}
}

type p2cPicker struct {
	pickerAgl *loadAggregate
}

func (p2c *p2cPicker) Pick(ctx context.Context, info balancer.PickInfo) (conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	item, done := p2c.pickerAgl.Next()
	if item == nil {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}

	return item.(balancer.SubConn), done, nil
}
