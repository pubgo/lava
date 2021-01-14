package p2c

import (
	"context"
	"github.com/pubgo/golug/golug_balancer/xalg"
	"github.com/pubgo/golug/golug_balancer/xalg/p2c"

	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

type builder struct{}

func (t *builder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	npa := p2c.NewP2cAgl()
	for _, subConn := range readySCs {
		npa.Add(subConn)
	}

	return &picker{
		pickerAgl: npa,
	}
}

type picker struct {
	pickerAgl xalg.P2c
}

func (p2c *picker) Pick(ctx context.Context, info balancer.PickInfo) (
	conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	item, done := p2c.pickerAgl.Next()
	if item == nil {
		return nil, nil, balancer.ErrNoSubConnAvailable
	}
	return item.(balancer.SubConn), done, nil
}
