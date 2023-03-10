package mdns

import (
	"context"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/typex"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/pkg/proto/event/v1"
)

var _ registry.Watcher = (*Watcher)(nil)

func newWatcher(m *mdnsRegistry, service string, opt ...registry.WatchOpt) result.Result[registry.Watcher] {
	assert.If(service == "", "[service] should not be null")

	var allNodes typex.SyncMap
	var s = m.GetService(service)
	if s.IsErr() {
		return result.Err[registry.Watcher](s.Err())
	}

	for _, ss := range m.GetService(service).Unwrap() {
		for _, n := range ss.Nodes {
			allNodes.Set(n.Id, n)
		}
	}

	var ttl = m.cfg.TTL
	if ttl == 0 {
		ttl = time.Second * 30
	}

	results := make(chan *registry.Result)
	return result.OK[registry.Watcher](&Watcher{m: m, results: results, cancel: async.GoCtx(func(ctx context.Context) error {
		var fn = func() {
			defer recovery.Recovery(func(err error) {
				m.log.Err(err).Msg("watcher error")
			})

			m.log.Info().Str("service", service).Str("interval", ttl.String()).Msg("[mdns] registry watcher")

			var nodes typex.SyncMap
			var ss = m.GetService(service).Unwrap()
			for _, s := range ss {
				for _, n := range s.Nodes {
					allNodes.Set(n.Id, n)
				}
			}

			assert.Must(nodes.Each(func(id string, n *service.Node) {
				if allNodes.Has(id) {
					return
				}

				allNodes.Set(id, n)
				results <- &registry.Result{
					Action:  eventpbv1.EventType_UPDATE,
					Service: &service.Service{Name: service, Nodes: service.Nodes{n}},
				}
			}))

			assert.Must(allNodes.Each(func(id string, n *service.Node) {
				if nodes.Has(id) {
					return
				}

				allNodes.Delete(id)
				results <- &registry.Result{
					Action:  eventpbv1.EventType_DELETE,
					Service: &service.Service{Name: service, Nodes: service.Nodes{n}},
				}
			}))
		}

		for range time.Tick(ttl) {
			fn()
		}

		return nil
	})})
}

type Watcher struct {
	m       *mdnsRegistry
	results chan *registry.Result
	cancel  context.CancelFunc
}

func (m *Watcher) Next() result.Result[*registry.Result] {
	r, ok := <-m.results
	if !ok {
		return result.Wrap(r, registry.ErrWatcherStopped)
	}
	return result.OK(r)
}

func (m *Watcher) Stop() error {
	close(m.results)
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}
