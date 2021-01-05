package p2c

import (
	"context"
	"fmt"
	"math"
	"math/rand"
	"sync"
	"time"

	"go.uber.org/atomic"
	"google.golang.org/grpc/balancer"
	"google.golang.org/grpc/balancer/base"
	"google.golang.org/grpc/resolver"
)

const (
	Name            = "p2c_ewma"
	decayTime       = int64(time.Second * 10) // default value from finagle
	forcePick       = int64(time.Second)
	initSuccess     = 1000
	throttleSuccess = initSuccess / 2
	penalty         = int64(math.MaxInt32)
	pickTimes       = 3
	logInterval     = time.Minute
)

func init() {
	balancer.Register(newBuilder())
}

type p2cPickerBuilder struct {
}

func newBuilder() balancer.Builder {
	return base.NewBalancerBuilder(Name, new(p2cPickerBuilder))
}

func (b *p2cPickerBuilder) Build(readySCs map[resolver.Address]balancer.SubConn) balancer.Picker {
	if len(readySCs) == 0 {
		return base.NewErrPicker(balancer.ErrNoSubConnAvailable)
	}

	var conns []*subConn
	for addr, conn := range readySCs {
		conns = append(conns, &subConn{
			addr:    addr,
			conn:    conn,
			success: *atomic.NewUint64(initSuccess),
		})
	}

	return &p2cPicker{
		conns: conns,
		r:     rand.New(rand.NewSource(time.Now().UnixNano())),
		stamp: atomic.NewDuration(0),
	}
}

type p2cPicker struct {
	conns []*subConn
	r     *rand.Rand
	stamp *atomic.Duration
	lock  sync.Mutex
}

func (p *p2cPicker) Pick(ctx context.Context, info balancer.PickInfo) (
	conn balancer.SubConn, done func(balancer.DoneInfo), err error) {
	p.lock.Lock()
	defer p.lock.Unlock()

	var chosen *subConn
	switch len(p.conns) {
	case 0:
		return nil, nil, balancer.ErrNoSubConnAvailable
	case 1:
	case 2:
	default:
		var node1, node2 *subConn
		for i := 0; i < pickTimes; i++ {
			a := p.r.Intn(len(p.conns))
			b := p.r.Intn(len(p.conns) - 1)
			if b >= a {
				b++
			}
			node1 = p.conns[a]
			node2 = p.conns[b]
			if node1.healthy() && node2.healthy() {
				break
			}
		}

	}
	return chosen.conn, nil, nil
}

func (p *p2cPicker) logStats() {
	var stats []string

	p.lock.Lock()
	defer p.lock.Unlock()

	for _, conn := range p.conns {
		stats = append(stats, fmt.Sprintf("conn: %s, load: %d, reqs: %d",
			conn.addr.Addr, conn.load(), conn.requests.Swap(0)))
	}
}

type subConn struct {
	addr     resolver.Address
	conn     balancer.SubConn
	lag      atomic.Uint64
	inflight atomic.Int64
	success  atomic.Uint64
	requests atomic.Int64
	last     int64
	pick     int64
}

func (c *subConn) healthy() bool {
	return c.success.Load() > throttleSuccess
}

func (c *subConn) load() int64 {
	// plus one to avoid multiply zero
	lag := int64(math.Sqrt(float64(c.lag.Load() + 1)))
	load := lag * (c.inflight.Load() + 1)
	if load == 0 {
		return penalty
	} else {
		return load
	}
}
