package signaling

import (
	"context"
	"sync"
)

// Multiplex 在单个 Broker Recv 循环上 fan-out 信令，避免多个 ICE 会话争抢 Recv。
type Multiplex struct {
	inner  Broker
	selfID string

	mu      sync.Mutex
	subs    map[uint64]chan Message
	nextID  uint64
	started bool
	stopCh  chan struct{}
	stopOnce sync.Once
	wg      sync.WaitGroup
}

// NewMultiplex 包装底层 Broker；每个 ICE 会话通过 Session 获得独立 Recv 通道。
func NewMultiplex(inner Broker, selfID string) *Multiplex {
	return &Multiplex{
		inner:  inner,
		selfID: selfID,
		subs:   make(map[uint64]chan Message),
		stopCh: make(chan struct{}),
	}
}

// Session 创建一条 ICE 信令会话；Close 后不再接收消息。
func (m *Multiplex) Session() *SessionBroker {
	m.ensureDispatcher()
	m.mu.Lock()
	id := m.nextID
	m.nextID++
	ch := make(chan Message, 64)
	m.subs[id] = ch
	m.mu.Unlock()
	return &SessionBroker{m: m, id: id, ch: ch}
}

// Close 停止分发（不关闭底层 Broker，由调用方负责）。
func (m *Multiplex) Close() error {
	m.stopOnce.Do(func() { close(m.stopCh) })
	m.wg.Wait()
	return nil
}

func (m *Multiplex) ensureDispatcher() {
	m.mu.Lock()
	if m.started {
		m.mu.Unlock()
		return
	}
	m.started = true
	m.mu.Unlock()

	m.wg.Add(1)
	go m.dispatchLoop()
}

func (m *Multiplex) dispatchLoop() {
	defer m.wg.Done()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		<-m.stopCh
		cancel()
	}()
	for {
		msg, err := m.inner.Recv(ctx, m.selfID)
		if err != nil {
			return
		}
		m.broadcast(msg)
	}
}

func (m *Multiplex) broadcast(msg Message) {
	m.mu.Lock()
	subs := make([]chan Message, 0, len(m.subs))
	for _, ch := range m.subs {
		subs = append(subs, ch)
	}
	m.mu.Unlock()

	for _, ch := range subs {
		select {
		case ch <- msg:
		default:
			select {
			case <-ch:
			default:
			}
			select {
			case ch <- msg:
			case <-m.stopCh:
				return
			}
		}
	}
}

func (m *Multiplex) unsubscribe(id uint64) {
	m.mu.Lock()
	if ch, ok := m.subs[id]; ok {
		delete(m.subs, id)
		close(ch)
	}
	m.mu.Unlock()
}

// SessionBroker 是单条 ICE 协商使用的 Broker 视图。
type SessionBroker struct {
	m  *Multiplex
	id uint64
	ch chan Message
}

func (s *SessionBroker) Send(ctx context.Context, msg Message) error {
	return s.m.inner.Send(ctx, msg)
}

func (s *SessionBroker) Recv(ctx context.Context, selfID string) (Message, error) {
	select {
	case msg, ok := <-s.ch:
		if !ok {
			return Message{}, context.Canceled
		}
		return msg, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	case <-s.m.stopCh:
		return Message{}, context.Canceled
	}
}

func (s *SessionBroker) Close() error {
	s.m.unsubscribe(s.id)
	return nil
}

var (
	_ Broker = (*SessionBroker)(nil)
)
