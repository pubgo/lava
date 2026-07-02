package signaling

import (
	"context"
	"sync"
)

// MemoryBroker 用于单测/PoC 的内存信令交换（双端各持对端引用）。
type MemoryBroker struct {
	mu   sync.Mutex
	ch   map[string]chan Message
	done chan struct{}
}

func NewMemoryBroker() *MemoryBroker {
	return &MemoryBroker{
		ch:   make(map[string]chan Message),
		done: make(chan struct{}),
	}
}

func (m *MemoryBroker) register(id string) chan Message {
	m.mu.Lock()
	defer m.mu.Unlock()
	if c, ok := m.ch[id]; ok {
		return c
	}
	c := make(chan Message, 64)
	m.ch[id] = c
	return c
}

func (m *MemoryBroker) Send(ctx context.Context, msg Message) error {
	if msg.To == "" {
		return nil
	}
	dst := m.register(msg.To)
	select {
	case dst <- msg:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	case <-m.done:
		return context.Canceled
	}
}

func (m *MemoryBroker) Recv(ctx context.Context, selfID string) (Message, error) {
	in := m.register(selfID)
	select {
	case msg := <-in:
		return msg, nil
	case <-ctx.Done():
		return Message{}, ctx.Err()
	case <-m.done:
		return Message{}, context.Canceled
	}
}

func (m *MemoryBroker) Close() error {
	select {
	case <-m.done:
	default:
		close(m.done)
	}
	return nil
}

// Pair 返回两个互通的 Broker 端点（同一底层交换）。
func Pair() (a, b *MemoryBroker) {
	hub := NewMemoryBroker()
	return hub, hub
}
