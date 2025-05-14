package p2c

import (
	"github.com/pubgo/funk/errors"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

var _ base.PickerBuilder = (*p2cBalancer)(nil)

type p2cBalancer struct{}

func (p2c *p2cBalancer) Build(info base.PickerBuildInfo) balancer.Picker {
	if len(info.ReadySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	// 创建一个新的负载均衡器
	connM := newP2cAgl()
	for subConn := range info.ReadySCs {
		connM.Add(subConn)
	}

	return &p2cPicker{pickerAgl: connM}
}

var _ balancer.Picker = (*p2cPicker)(nil)

type p2cPicker struct {
	pickerAgl *loadAggregate
}

func (p2c *p2cPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	// TODO 负载策略
	// info 可以根据具体的method做负载
	// 可以根据context的value(可以是userID等, 或者权重)做负载

	item, done := p2c.pickerAgl.Next(info)
	if item == nil {
		return balancer.PickResult{}, errors.Wrap(balancer.ErrNoSubConnAvailable, "p2c pick error, no SubConn is available")
	}

	return balancer.PickResult{SubConn: item.(balancer.SubConn), Done: done}, nil
}
