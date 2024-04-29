package discovery

import (
	"context"

	"github.com/pubgo/funk/result"

	"github.com/pubgo/lava/core/service"
)

func NewNoopDiscovery() Discovery {
	return new(noopDiscovery)
}

var (
	_ Discovery = (*noopDiscovery)(nil)
	_ Watcher   = (*noopDiscovery)(nil)
)

type noopDiscovery struct{}

func (n *noopDiscovery) Next() result.Result[*Result] {
	return result.Err[*Result](ErrWatcherStopped)
}

func (n *noopDiscovery) Stop() error { return nil }

func (n *noopDiscovery) String() string { return "noop" }

func (n *noopDiscovery) Watch(ctx context.Context, srv string, opts ...WatchOpt) result.Result[Watcher] {
	return result.OK[Watcher](n)
}

func (n *noopDiscovery) GetService(ctx context.Context, srv string, opts ...GetOpt) result.Result[[]*service.Service] {
	return result.Result[[]*service.Service]{}
}
