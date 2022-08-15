package mdns

import (
	"context"
	"time"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/xerr"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/gen/event/eventpbv1"
	"github.com/pubgo/lava/internal/pkg/syncx"
	"github.com/pubgo/lava/internal/pkg/typex"
)

var _ registry.Watcher = (*Watcher)(nil)

func newWatcher(m *mdnsRegistry, service string, opt ...registry.WatchOpt) *Watcher {
	assert.If(service == "", "[service] should not be null")

	var allNodes typex.SMap
	services := assert.Must1(m.GetService(service))
	for i := range services {
		for _, n := range services[i].Nodes {
			allNodes.Set(n.Id, n)
		}
	}

	var ttl = m.cfg.TTL
	if ttl == 0 {
		ttl = time.Second * 30
	}

	results := make(chan *registry.Result)
	return &Watcher{m: m, results: results, cancel: syncx.GoCtx(func(ctx context.Context) {
		var fn = func() {
			defer recovery.Recovery(func(err xerr.XErr) {
				m.log.WithErr(err).Error("watcher error")
			})

			m.log.With(typex.M{
				"service":  service,
				"interval": ttl.String(),
			}).Info("[mdns] registry watcher")

			var nodes typex.SMap
			serviceList := assert.Must1(m.GetService(service))
			for i := range serviceList {
				for _, n := range serviceList[i].Nodes {
					nodes.Set(n.Id, n)
				}
			}

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
	})}
}

type Watcher struct {
	m       *mdnsRegistry
	results chan *registry.Result
	cancel  context.CancelFunc
}

func (m *Watcher) Next() (*registry.Result, error) {
	result, ok := <-m.results
	if !ok {
		return nil, registry.ErrWatcherStopped
	}

	return result, nil
}

func (m *Watcher) Stop() error {
	close(m.results)
	if m.cancel != nil {
		m.cancel()
	}
	return nil
}
