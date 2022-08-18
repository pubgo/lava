package mdns

import (
	"context"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/gen/proto/event/v1"
	"github.com/pubgo/lava/internal/pkg/result"
	"github.com/pubgo/lava/internal/pkg/syncx"
	"github.com/pubgo/lava/internal/pkg/typex"
)

var _ registry.Watcher = (*Watcher)(nil)

func newWatcher(m *mdnsRegistry, service string, opt ...registry.WatchOpt) result.Result[registry.Watcher] {
	assert.If(service == "", "[service] should not be null")

	var ret result.Result[registry.Watcher]

	var allNodes typex.SMap
	for _, s := range m.GetService(service) {
		if s.IsErr() {
			return ret.WithErr(s.Err())
		}

		for _, n := range s.Get().Nodes {
			allNodes.Set(n.Id, n)
		}
	}

	var ttl = m.cfg.TTL
	if ttl == 0 {
		ttl = time.Second * 30
	}

	results := make(chan *registry.Result)
	return ret.WithVal(&Watcher{m: m, results: results, cancel: syncx.GoCtx(func(ctx context.Context) {
		var fn = func() {
			defer recovery.Recovery(func(err xerr.XErr) {
				m.log.WithErr(err).Error("watcher error")
			})

			m.log.With(typex.M{
				"service":  service,
				"interval": ttl.String(),
			}).Info("[mdns] registry watcher")

			var nodes typex.SMap
			m.GetService(service).Range(func(r result.Result[*registry.Service]) {
				for _, n := range r.Get().Nodes {
					allNodes.Set(n.Id, n)
				}
			})

			assert.Must(nodes.Each(func(id string, n *registry.Node) {
				if allNodes.Has(id) {
					return
				}

				allNodes.Set(id, n)
				results <- &registry.Result{
					Action:  eventpbv1.EventType_UPDATE,
					Service: &registry.Service{Name: service, Nodes: registry.Nodes{n}},
				}
			}))

			assert.Must(allNodes.Each(func(id string, n *registry.Node) {
				if nodes.Has(id) {
					return
				}

				allNodes.Delete(id)
				results <- &registry.Result{
					Action:  eventpbv1.EventType_DELETE,
					Service: &registry.Service{Name: service, Nodes: registry.Nodes{n}},
				}
			}))
		}

		for range time.Tick(ttl) {
			fn()
		}
	})})
}

type Watcher struct {
	m       *mdnsRegistry
	results chan *registry.Result
	cancel  context.CancelFunc
}

func (m *Watcher) Next() (*registry.Result, error) {
	r, ok := <-m.results
	if !ok {
		return nil, registry.ErrWatcherStopped
	}
	return r, nil
}

func (m *Watcher) Stop() error {
	close(m.results)
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}
