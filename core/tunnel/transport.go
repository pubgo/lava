package tunnel

import (
	"context"
	"fmt"
	"sync"
)

// 支持的传输协议常量
const (
	// TransportYamux yamux 传输协议
	TransportYamux = "yamux"
	// TransportQUIC QUIC 传输协议
	TransportQUIC = "quic"
	// TransportHTTP HTTP CONNECT 传输协议
	TransportHTTP = "http"
	// TransportKCP KCP 传输协议
	TransportKCP = "kcp"
)

var (
	transportMu        sync.RWMutex
	transportFactories = make(map[string]TransportFactory)
)

// TransportFactory 传输层工厂函数
type TransportFactory func(opts *TransportOptions) (Transport, error)

// RegisterTransport 注册传输层工厂
func RegisterTransport(name string, factory TransportFactory) {
	transportMu.Lock()
	defer transportMu.Unlock()
	if factory == nil {
		panic("tunnel: RegisterTransport factory is nil")
	}
	if _, dup := transportFactories[name]; dup {
		panic("tunnel: RegisterTransport called twice for factory " + name)
	}
	transportFactories[name] = factory
}

// NewTransport 创建传输层实例
func NewTransport(name string, opts *TransportOptions) (Transport, error) {
	transportMu.RLock()
	factory, ok := transportFactories[name]
	transportMu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("%w: %s", ErrTransportNotSupported, name)
	}

	if opts == nil {
		opts = DefaultTransportOptions()
	}

	return factory(opts)
}

// GetTransportFactory 获取已注册的传输层工厂
func GetTransportFactory(name string) (TransportFactory, bool) {
	transportMu.RLock()
	defer transportMu.RUnlock()
	f, ok := transportFactories[name]
	return f, ok
}

// ListTransports 列出所有已注册的传输层
func ListTransports() []string {
	transportMu.RLock()
	defer transportMu.RUnlock()

	names := make([]string, 0, len(transportFactories))
	for name := range transportFactories {
		names = append(names, name)
	}
	return names
}

// MustNewTransport 创建传输层实例，失败时 panic
func MustNewTransport(name string, opts *TransportOptions) Transport {
	t, err := NewTransport(name, opts)
	if err != nil {
		panic(err)
	}
	return t
}

// DialTransport 使用指定传输协议连接到服务端
func DialTransport(ctx context.Context, transportName, addr string, opts *TransportOptions) (Session, error) {
	t, err := NewTransport(transportName, opts)
	if err != nil {
		return nil, err
	}
	return t.Dial(ctx, addr)
}

// ListenTransport 使用指定传输协议监听
func ListenTransport(ctx context.Context, transportName, addr string, opts *TransportOptions) (Listener, error) {
	t, err := NewTransport(transportName, opts)
	if err != nil {
		return nil, err
	}
	return t.Listen(ctx, addr)
}
