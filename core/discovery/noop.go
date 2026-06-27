package discovery

import (
	"context"

	"github.com/pubgo/funk/v2/result"

	"github.com/pubgo/lava/v2/core/service"
)

// NewNoopDiscovery 返回一个不执行任何服务发现的 Discovery 实现。
func NewNoopDiscovery() Discovery {
	return new(noopDiscovery)
}

var (
	_ Discovery = (*noopDiscovery)(nil)
	_ Watcher   = (*noopDiscovery)(nil)
)

type noopDiscovery struct{}

func (n *noopDiscovery) Next() (r result.Result[*Result]) {
	return result.Fail[*Result](ErrWatcherStopped)
}

func (n *noopDiscovery) Stop() error { return nil }

func (n *noopDiscovery) String() string { return "noop" }

func (n *noopDiscovery) Watch(_ context.Context, _ string, _ ...WatchOpt) result.Result[Watcher] {
	return result.OK[Watcher](n)
}

func (n *noopDiscovery) GetService(_ context.Context, _ string, _ ...GetOpt) result.Result[[]*service.Service] {
	return result.OK([]*service.Service(nil))
}
