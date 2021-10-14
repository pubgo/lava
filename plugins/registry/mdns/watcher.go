package mdns

import (
	"context"
	"time"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
)

var _ registry.Watcher = (*Watcher)(nil)

func newWatcher(m *mdnsRegistry, service string, opt ...registry.WatchOpt) *Watcher {
	xerror.Assert(service == "", "[service] should not be null")

	var allNodes typex.SMap
	services, err := m.GetService(service)
	xerror.Panic(err)
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
	return &Watcher{results: results, cancel: fx.Tick(func(_ctx fx.Ctx) {
		defer xerror.Resp(func(err xerror.XErr) {
			logz.Named(Name).Desugar().Error("watcher error", logger.WithErr(err)...)
		})

		logz.Named(Name).Infof("[mdns] registry watch service(%s) on interval(%s)", service, ttl)

		var nodes typex.SMap
		services, err := m.GetService(service)
		xerror.PanicF(err, "Watch Service %s Error", service)
		for i := range services {
			for _, n := range services[i].Nodes {
				nodes.Set(n.Id, n)
			}
		}

		xerror.Panic(nodes.Each(func(id string, n *registry.Node) {
			if allNodes.Has(id) {
				return
			}

			allNodes.Set(id, n)
			results <- &registry.Result{
				Action:  types.EventType_UPDATE,
				Service: &registry.Service{Name: service, Nodes: registry.Nodes{n}},
			}
		}))

		xerror.Panic(allNodes.Each(func(id string, n *registry.Node) {
			if nodes.Has(id) {
				return
			}

			allNodes.Delete(id)
			results <- &registry.Result{
				Action:  types.EventType_DELETE,
				Service: &registry.Service{Name: service, Nodes: registry.Nodes{n}},
			}
		}))
	}, ttl)}
}

type Watcher struct {
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
