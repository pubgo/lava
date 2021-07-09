package p2c

import (
	"github.com/pubgo/x/q"
	"github.com/pubgo/xerror"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
)

var _ base.PickerBuilder = (*p2cBalancer)(nil)

type p2cBalancer struct{}

func (p2c *p2cBalancer) Build(info base.PickerBuildInfo) balancer.Picker {
	if info.ReadySCs == nil || len(info.ReadySCs) == 0 {
		return base.NewErrPicker(xerror.Wrap(balancer.ErrNoSubConnAvailable))
	}

	// 创建一个新的负载均衡器
	npa := NewP2cAgl()
	for subConn,info := range info.ReadySCs {
		q.Q(info)
		subConn.Connect()
		npa.Add(subConn)
	}

	return &p2cPicker{pickerAgl: npa}
}

var _ balancer.Picker = (*p2cPicker)(nil)

type p2cPicker struct {
	pickerAgl *loadAggregate
}

func (p2c *p2cPicker) Pick(info balancer.PickInfo) (balancer.PickResult, error) {
	item, done := p2c.pickerAgl.Next()
	if item == nil {
		return balancer.PickResult{}, xerror.Wrap(balancer.ErrNoSubConnAvailable)
	}

	return balancer.PickResult{SubConn: item.(balancer.SubConn), Done: done}, nil
}
