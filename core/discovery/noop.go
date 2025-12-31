package discovery

import (
	"context"

	"github.com/pubgo/lava/v2/core/service"

	"github.com/pubgo/funk/v2/result"
)

func NewNoopDiscovery() Discovery {
	return new(noopDiscovery)
}

var (
	_ Discovery = (*noopDiscovery)(nil)
	_ Watcher   = (*noopDiscovery)(nil)
)

type noopDiscovery struct{}

func (n *noopDiscovery) Next() (r result.Result[*Result]) {
	return r.WithErr(ErrWatcherStopped)
}

func (n *noopDiscovery) Stop() error { return nil }

func (n *noopDiscovery) String() string { return "noop" }

func (n *noopDiscovery) Watch(ctx context.Context, srv string, opts ...WatchOpt) result.Result[Watcher] {
	return result.OK[Watcher](n)
}

func (n *noopDiscovery) GetService(ctx context.Context, srv string, opts ...GetOpt) result.Result[[]*service.Service] {
	return result.Result[[]*service.Service]{}
}
