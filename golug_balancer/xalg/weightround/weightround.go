package weightround

import (
	"sync"

	"github.com/pubgo/golug/golug_balancer/xalg"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc/balancer"
)

//权重轮询算法，后续可能作为一个balancer算法使用。当前先实现算法
const defaultWRCap = 8

type WR struct {
	rmMux sync.RWMutex
	items []*WrNode
}

type WrNode struct {
	name      string
	current   int64
	effective int64
	wight     int64
	item      interface{}
}

func NewWr() xalg.WeightRound {
	return &WR{
		items: make([]*WrNode, 0, defaultWRCap),
	}
}

//允许有同名的出现，考虑到增加负载能力的时候可能传入多个同名service
func (wr *WR) Add(name string, weight int64, item interface{}) {
	wr.rmMux.Lock()
	defer wr.rmMux.Unlock()
	wrn := &WrNode{
		name:      name,
		current:   weight,
		effective: weight,
		item:      item,
	}

	wr.items = append(wr.items, wrn)
}

// return may be nil
func (wr *WR) Next() (interface{}, func(info balancer.DoneInfo)) {
	wr.rmMux.Lock()
	defer wr.rmMux.Unlock()
	if len(wr.items) == 0 || wr.items == nil {
		xlog.Info("wr items length = 0 or items=nil ")
		return nil, func(info balancer.DoneInfo) {}
	}

	var next *WrNode
	var total int64
	for _, item := range wr.items {
		total += item.effective
		item.current += item.effective
		if item.effective < item.wight {
			item.effective++
		}

		if next == nil || next.current < item.current {
			next = item
		}
	}

	if next != nil {
		next.current -= total
	}

	return next.item, func(info balancer.DoneInfo) {}
}
