package gossip

import (
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/plugins/registry/registry_type"
)

type gossipWatcher struct {
	wo   registry_type.WatchOpts
	next chan *registry_type.Result
	stop chan bool
}

func newGossipWatcher(ch chan *registry_type.Result, stop chan bool, opts ...registry_type.WatchOpt) (registry_type.Watcher, error) {
	var wo registry_type.WatchOpts
	for _, o := range opts {
		o(&wo)
	}

	return &gossipWatcher{
		wo:   wo,
		next: ch,
		stop: stop,
	}, nil
}

func (m *gossipWatcher) Next() (*registry_type.Result, error) {
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

			nr := &registry_type.Result{}
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
