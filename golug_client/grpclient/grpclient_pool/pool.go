package grpclient_pool

import (
	"sync"
	"time"

	"google.golang.org/grpc"
)

const (
	// 偿试清理间隔
	cleanupInterval = time.Minute * 10
	minTTL = 1 // second
	minPoolSize = 1
)

type Pool struct {
	size int
	ttl  int64

	sync.Mutex
	connList map[string]*poolManager
}

type ClientConn struct {
	*grpc.ClientConn

	// 创建状态
	newCreated bool
	created    time.Time

	// 使用中的引用计数
	refCount int64

	// 用来决定是否关闭
	closable bool
}

func NewPool(size int, ttl time.Duration) *Pool {
	ttls := int64(ttl.Seconds())
	if ttls < 0 {
		ttls = minTTL
	}

	if size < 0 {
		size = minPoolSize
	}

	pool := &Pool{
		size:     size,
		ttl:      ttls,
		connList: make(map[string]*poolManager),
	}

	go pool.cleanup()

	return pool
}

// Init modify the pool configurations.
// This configurations will represent on the managers create after this action.
func (p *Pool) Init(size int, ttl time.Duration) {
	p.Lock()
	defer p.Unlock()

	ttls := int64(ttl.Seconds())
	if ttls > 0 && ttls != p.ttl {
		p.ttl = ttls
	}

	if size > 0 && size != p.size {
		p.size = size
	}
}

func (p *Pool) GetConn(addr string, opts ...grpc.DialOption) (*ClientConn, error) {
	return p.getManager(addr).get(opts...)
}

func (p *Pool) Release(addr string, conn *ClientConn, err error) {
	// otherwise put it back for reuse
	p.getManager(addr).put(conn, err)
}

func (p *Pool) getManager(addr string) *poolManager {
	p.Lock()
	defer p.Unlock()

	manager := p.connList[addr]
	if manager == nil {
		manager = newManager(addr, p.size, p.ttl)
		p.connList[addr] = manager
	}

	return manager
}

func (p *Pool) cleanup() {
	timer := time.NewTicker(cleanupInterval)
	for range timer.C {
		cleans := p.findCanCleanups()

		if len(cleans) == 0 {
			continue
		}

		for _, manager := range cleans {
			manager.cleanup()
		}
	}
}

func (p *Pool) findCanCleanups() map[string]*poolManager {
	p.Lock()
	defer p.Unlock()

	cleans := make(map[string]*poolManager)

	for addr, manager := range p.connList {
		if manager != nil && manager.canCleanup() {
			cleans[addr] = manager
			delete(p.connList, addr)
		}
	}

	return cleans
}
