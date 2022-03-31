package gossip

import (
	"github.com/pubgo/lava/core/registry"
)

type gossipWatcher struct {
	wo   registry.WatchOpts
	next chan *registry.Result
	stop chan bool
}

func newGossipWatcher(ch chan *registry.Result, stop chan bool, opts ...registry.WatchOpt) (registry.Watcher, error) {
	var wo registry.WatchOpts
	for _, o := range opts {
		o(&wo)
	}

	return &gossipWatcher{
		wo:   wo,
		next: ch,
		stop: stop,
	}, nil
}

func (m *gossipWatcher) Next() (*registry.Result, error) {
	for {
		select {
		case r, ok := <-m.next:
			if !ok {
				return nil, registry.ErrWatcherStopped
			}

			// check watch opts
			if len(m.wo.Service) > 0 && r.Service.Name != m.wo.Service {
				continue
			}

			nr := &registry.Result{}
			*nr = *r
			return nr, nil
		case <-m.stop:
			return nil, registry.ErrWatcherStopped
		}
	}
}

func (m *gossipWatcher) Stop() error {
	select {
	case <-m.stop:
		return nil
	default:
		close(m.stop)
	}
	return nil
}
