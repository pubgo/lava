package p2c

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"google.golang.org/grpc/balancer"
)

type node struct {
	item      interface{}
	loadCount int64
}

// 最小连接负载均衡策略
type loadAggregate struct {
	items []*node
	mu    sync.Mutex
	rand  *rand.Rand
}

func NewP2cAgl() *loadAggregate {
	return &loadAggregate{
		items: make([]*node, 0),
		rand:  rand.New(rand.NewSource(time.Now().Unix())),
	}
}

func (la *loadAggregate) Add(n interface{}) {
	la.items = append(la.items, &node{item: n})
}

func (la *loadAggregate) Next(info balancer.PickInfo) (interface{}, func(info balancer.DoneInfo)) {
	var election, alternative *node
	switch len(la.items) {
	case 0:
		return nil, func(info balancer.DoneInfo) {}
	case 1:
		election = la.items[0]
	default:
		// 随机获取两个连接
		la.mu.Lock()
		x := la.rand.Intn(len(la.items))
		y := la.rand.Intn(len(la.items) - 1)
		la.mu.Unlock()
		if x == y {
			y += 1
		}

		// 获取连接中使用数最小的那个
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
