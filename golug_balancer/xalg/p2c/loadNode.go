package p2c

import (
	"github.com/pubgo/golug/golug_balancer/xalg"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/balancer"
)

type loadNode struct {
	item      interface{}
	loadCount int64
}

type loadAggregate struct {
	items []*loadNode
	mu    sync.Mutex
	rand  *rand.Rand
}

func NewP2cAgl() xalg.P2c {
	return &loadAggregate{
		items: make([]*loadNode, 0),
		rand:  rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (la *loadAggregate) Add(item interface{}) {
	la.items = append(la.items, &loadNode{item: item})
}

func (la *loadAggregate) Next() (interface{}, func(info balancer.DoneInfo)) {
	var election, alternative *loadNode
	switch len(la.items) {
	case 0:
		return nil, func(info balancer.DoneInfo) {}
	case 1:
		election = la.items[0]
	default:
		la.mu.Lock()
		x := la.rand.Intn(len(la.items))
		y := la.rand.Intn(len(la.items) - 1)
		la.mu.Unlock()
		if x == y {
			y += 1
		}
		election, alternative = la.items[x], la.items[y]
		if election.loadCount > alternative.loadCount {
			election = alternative
		}
	}

	atomic.AddInt64(&election.loadCount, 1)

	return election.item, func(info balancer.DoneInfo) {
		atomic.AddInt64(&election.loadCount, -1)
	}
}
