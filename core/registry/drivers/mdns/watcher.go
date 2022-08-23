package mdns

import (
	"context"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/syncx"
	"github.com/pubgo/funk/xerr"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/gen/proto/event/v1"
	"github.com/pubgo/lava/internal/pkg/typex"
)

var _ registry.Watcher = (*Watcher)(nil)

func newWatcher(m *mdnsRegistry, service string, opt ...registry.WatchOpt) result.Result[registry.Watcher] {
	assert.If(service == "", "[service] should not be null")

	var allNodes typex.SMap
	for _, s := range m.GetService(service) {
		if s.IsErr() {
			return result.Err[registry.Watcher](s.Err())
		}

		for _, n := range s.Unwrap().Nodes {
			allNodes.Set(n.Id, n)
		}
	}

	var ttl = m.cfg.TTL
	if ttl == 0 {
		ttl = time.Second * 30
	}

	results := make(chan *registry.Result)
	return result.OK[registry.Watcher](&Watcher{m: m, results: results, cancel: syncx.GoCtx(func(ctx context.Context) result.Error {
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
				if r.IsNil() {
					return
				}

				for _, n := range r.Unwrap().Nodes {
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

		return result.Error{}
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
