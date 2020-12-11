package grpclient_pool

import (
	"sync"
	"time"
	"unsafe"

	"google.golang.org/grpc"
	"google.golang.org/grpc/connectivity"
)

const (
	// 闲时每个连接处理的请求
	requestPerConn = 8
	// 多久可以清理连接
	cleanupDuration = time.Minute * 30
	// 随机打散的范围
	randCreatedIn = 60
)

type poolManager struct {
	sync.Mutex

	// 管理当前可选的连接
	indexes []*ClientConn
	// 真实可用的连接
	data map[*ClientConn]struct{}
	// 真实 Pool 的大小，也就是最大池子长度
	size int
	// 连接有效期
	ttl int64
	// 连接地址
	addr string
	// 最后使用时间
	updatedAt time.Time
	// 创建连接的门票
	tickets ticket
}

type connState uint8

const (
	invalid connState = iota
	poolFull
	lessThanRef
	moreThanRef
)

func newManager(addr string, size int, ttl int64) *poolManager {
	tickerCount := size
	manager := &poolManager{
		indexes: make([]*ClientConn, 0, 0),
		data:    make(map[*ClientConn]struct{}),
		size:    size,
		ttl:     ttl,
		addr:    addr,
		tickets: newTicker(tickerCount),
	}

	return manager
}

func (m *poolManager) isValid(conn *ClientConn) connState {
	// 无效 1
	if conn == nil {
		return invalid
	}

	// 无效 2
	// 已经从真实数据中移除
	if _, ok := m.data[conn]; !ok {
		return invalid
	}

	// 无效 3
	// 已经过期
	if int64(time.Since(conn.created).Seconds()) >= m.ttl {
		return invalid
	}

	// 无效 4
	// 如果连接状态都不可用了则直接是无效连接
	if conn.GetState() != connectivity.Ready {
		return invalid
	}

	// ===================================================================
	// 有效 1
	// 如果池子已经到达了上限了，那么只要此连接状态是可用的，就认为是有效的连接
	if len(m.data) >= m.size {
		return poolFull
	}
	// 有效 2
	// 如果池子还没有满，但是当前连接已经到达了每个连接上承载的请求数时，认为些连接无效
	if conn.refCount < requestPerConn {
		return lessThanRef
	} else {
		return moreThanRef
	}
	// ===================================================================

	// 其它情况均为无效
	return invalid
}

func (m *poolManager) get(opts ...grpc.DialOption) (*ClientConn, error) {
	conn, found := m.tryFindOne()
	if found {
		return conn, nil
	}

	// 获取门票
	m.tickets.get()

	// 1. 再次确认是否已经真的没有连接可用了
	conn, found = m.tryFindOne()
	if found {
		// 老连接直接放票
		// 如果是新建连接则要等处理完一次请求后再放票
		m.tickets.release()
		return conn, nil
	}

	// 2. 真的没有连接可用的时间，这时新建连接
	return m.create(opts...)
}

// tryFindOne 偿试找一个可用的连接
func (m *poolManager) tryFindOne() (*ClientConn, bool) {
	m.Lock()
	defer m.Unlock()
	m.updatedAt = time.Now() // 更新最后使用时间

	var nextRound []*ClientConn // 进入下一轮
	for idx, conn := range m.indexes {
		// 直接移除，下一个请求直接不再选择
		// 标识可关闭
		state := m.isValid(conn)
		if state == invalid {
			conn.closable = true
			m.tryClose(conn)
			continue
		} else if state == moreThanRef {
			// 此连接参与下一轮服务
			nextRound = append(nextRound, conn)
			continue
		}

		// 此连接参与下一轮服务
		nextRound = append(nextRound, conn)

		// 在返回前处理下索引对象，将其移动到最末尾
		m.indexes = m.indexes[idx+1:]
		m.indexes = append(m.indexes, nextRound...)

		// 增加当前连接的引用计数
		conn.refCount++

		return conn, true
	}

	// 找不到时则认为只有存在于下一轮中的连接可用
	m.indexes = nextRound

	return nil, false
}

func (m *poolManager) create(opts ...grpc.DialOption) (*ClientConn, error) {
	// 如果从 manager 中找不到可用的连接时，新建一个连接
	cc, err := grpc.Dial(m.addr, WithDefaultDialOptions(opts...)...)
	if err != nil {
		return nil, err
	}

	conn := &ClientConn{
		ClientConn: cc,
		newCreated: true,
		created:    m.backoff(cc), // 打乱时长
		refCount:   1,
		closable:   false,
	}


	return conn, nil
}

// backoff 计算一个连接平均值，防止一直在某个时间点全部失效
func (m *poolManager) backoff(cc *grpc.ClientConn) time.Time {
	p := uint(uintptr(unsafe.Pointer(cc))) % randCreatedIn


	return time.Now().Add(time.Second * time.Duration(p))
}

func (m *poolManager) put(conn *ClientConn, err error) {
	m.Lock()
	defer m.Unlock()

	if conn == nil {
		return
	}

	if conn.newCreated {
		conn.newCreated = false
		m.tickets.release()
	}

	conn.refCount--

	// 1. 任何一次请求错误就从连接池中移除，并且标识可关闭
	if err != nil {
		conn.closable = true
	}

	// 2.
	// 如果池子里已经满了则应该可关闭
	// 如果不在池子里，池子也不满则放入池子，并且开始复用
	_, inPool := m.data[conn]
	if !inPool {
		if len(m.data) >= m.size {
			conn.closable = true
		}
	}

	// 如果已经失败过，要执行后续处理
	if conn.closable {
		m.tryClose(conn)
		return
	}

	// 二次判断的原因是因为前面可能是 closable 的状态
	if !inPool {
		// 因为如果下一个选择还是会判断是否有效的，这里加入可提前让连接复用
		m.indexes = append(m.indexes, conn)
		m.data[conn] = struct{}{}
	}

	return
}

// tryClose 偿试关闭可关闭的连接
func (m *poolManager) tryClose(conn *ClientConn) {

	// 1. 从连接管理中移除，不让下一个选中
	delete(m.data, conn)

	// 2. 判断 refCount 是否为 0， 如果为 0 则直接关闭了
	if conn.refCount <= 0 {
		go func() {
			// 异步关闭，减少后面的请求阻塞
			if err := conn.ClientConn.Close(); err != nil {
			}
		}()
	}
}

// canCleanup 表示是否可以清理
// 独立于 cleanup 是让 pool 对 manager 的管理过程，没有锁间隙
func (m *poolManager) canCleanup() bool {
	m.Lock()
	defer m.Unlock()

	if time.Since(m.updatedAt) < cleanupDuration {
		return false
	}

	return true
}

// cleanup 定时清理无用的连接
func (m *poolManager) cleanup() {
	m.Lock()
	defer m.Unlock()

	for conn := range m.data {
		if conn.refCount > 0 {
		}
		_ = conn.ClientConn.Close()
	}
}

// ticket 门票概念
type ticket chan struct{}

func newTicker(size int) ticket {
	tks := make(ticket, size)

	for i := 0; i < size; i++ {
		tks <- struct{}{}
	}

	return tks
}

func (t ticket) get() {
	<-t
}

func (t ticket) release() {
	t <- struct{}{}
}

func (t ticket) size() int {
	return len(t)
}
