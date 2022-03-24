package mdns

import (
	"context"
	"time"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lava/event"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/plugins/registry/registry_type"
)

var logs = logging.Component(Name)
var _ registry_type.Watcher = (*Watcher)(nil)

func newWatcher(m *mdnsRegistry, service string, opt ...registry_type.WatchOpt) *Watcher {
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

	results := make(chan *registry_type.Result)
	return &Watcher{results: results, cancel: fx.Tick(func(_ctx fx.Ctx) {
		defer xerror.Resp(func(err xerror.XErr) {
			logs.WithErr(err).Error("watcher error")
		})

		logs.L().With(
			zap.String("service", service),
			zap.String("interval", ttl.String()),
		).Info("[mdns] registry watcher")

		var nodes typex.SMap
		services, err := m.GetService(service)
		xerror.PanicF(err, "Watch Service %s Error", service)
		for i := range services {
			for _, n := range services[i].Nodes {
				nodes.Set(n.Id, n)
			}
		}

		xerror.Panic(nodes.Each(func(id string, n *registry_type.Node) {
			if allNodes.Has(id) {
				return
			}

			allNodes.Set(id, n)
			results <- &registry_type.Result{
				Action:  event.EventType_UPDATE,
				Service: &registry_type.Service{Name: service, Nodes: registry_type.Nodes{n}},
			}
		}))

		xerror.Panic(allNodes.Each(func(id string, n *registry_type.Node) {
			if nodes.Has(id) {
				return
			}

			allNodes.Delete(id)
			results <- &registry_type.Result{
				Action:  event.EventType_DELETE,
				Service: &registry_type.Service{Name: service, Nodes: registry_type.Nodes{n}},
			}
		}))
	}, ttl)}
}

type Watcher struct {
	results chan *registry_type.Result
	cancel  context.CancelFunc
}

func (m *Watcher) Next() (*registry_type.Result, error) {
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
