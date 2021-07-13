// Package mdns is a multicast dns registry
package mdns

import (
	"context"
	"fmt"
	"time"

	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/registry"

	"github.com/grandcat/zeroconf"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	registry.Register(Name, NewWithMap)
}

func NewWithMap(m map[string]interface{}) (registry.Registry, error) {
	resolver, err := zeroconf.NewResolver()
	xerror.Panic(err, "Failed to initialize zeroconf resolver")

	var r = &mdnsRegistry{resolver: resolver}
	xerror.Panic(merge.MapStruct(&r.cfg, m))
	return r, nil
}

var _ registry.Registry = (*mdnsRegistry)(nil)
var _ registry.Watcher = (*mdnsRegistry)(nil)

type mdnsRegistry struct {
	cfg      Cfg
	services typex.SMap
	results  chan *registry.Result
	resolver *zeroconf.Resolver
	cancel   context.CancelFunc
}

func (m *mdnsRegistry) Next() (*registry.Result, error) {
	result, ok := <-m.results
	if !ok {
		return nil, registry.ErrWatcherStopped
	}

	return result, nil
}

func (m *mdnsRegistry) Stop() {
	close(m.results)
	if m.cancel != nil {
		m.cancel()
	}
}

func (m *mdnsRegistry) Register(service *registry.Service, optList ...registry.RegOpt) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(service == nil, "[service] should not be nil")
	xerror.Assert(len(service.Nodes) == 0, "[service] nodes should not be zero")

	node := service.Nodes[0]

	// 已经存在
	if m.services.Has(node.Id) {
		return nil
	}

	server, err := zeroconf.Register(node.Id, service.Name, "local.", node.GetPort(), []string{node.Id}, nil)
	xerror.PanicF(err, "[mdns] service %s register error", service.Name)

	var opts registry.RegOpts
	for i := range optList {
		optList[i](&opts)
	}

	m.services.Set(node.Id, server)
	return nil
}

func (m *mdnsRegistry) DeRegister(service *registry.Service, opt ...registry.DeRegOpt) (err error) {
	defer xerror.RespErr(&err)

	xerror.Assert(service == nil, "[service] should not be nil")
	xerror.Assert(len(service.Nodes) == 0, "[service] nodes should not be zero")

	node := service.Nodes[0]
	var val, ok = m.services.LoadAndDelete(node.Id)
	if !ok {
		return nil
	}

	val.(*zeroconf.Server).Shutdown()

	return nil
}

func (m *mdnsRegistry) GetService(name string, opts ...registry.GetOpt) (services []*registry.Service, _ error) {
	return services, xutil.Try(func() {
		entries := make(chan *zeroconf.ServiceEntry)
		_ = fx.Go(func(ctx context.Context) {
			for s := range entries {
				services = append(services, &registry.Service{
					Name: s.Service,
					Nodes: registry.Nodes{{
						Id:      s.Instance,
						Port:    s.Port,
						Address: fmt.Sprintf("%s:%d", s.AddrIPv4[0].String(), s.Port),
					}},
				})
			}
		})

		var gOpts registry.GetOpts
		for i := range opts {
			opts[i](&gOpts)
		}

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()

		xerror.PanicF(m.resolver.Browse(ctx, name, "local.", entries), "Failed to Lookup Service %s", name)
		<-ctx.Done()
	})
}

func (m *mdnsRegistry) ListServices(opts ...registry.ListOpt) (services []*registry.Service, _ error) {
	return services, nil
}

func (m *mdnsRegistry) Watch(service string, opt ...registry.WatchOpt) (registry.Watcher, error) {
	var watcher = &mdnsRegistry{results: make(chan *registry.Result)}

	return watcher, xutil.Try(func() {
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

		watcher.cancel = fx.Tick(func(_ctx fx.Ctx) {
			xlog.Infof("[mdns] registry watch service(%s) on interval(%s)", service, ttl)

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
				watcher.results <- &registry.Result{
					Action:  registry.Update.String(),
					Service: &registry.Service{Name: service, Nodes: registry.Nodes{n}},
				}
			}))

			xerror.Panic(allNodes.Each(func(id string, n *registry.Node) {
				if nodes.Has(id) {
					return
				}

				allNodes.Delete(id)
				watcher.results <- &registry.Result{
					Action:  registry.Delete.String(),
					Service: &registry.Service{Name: service, Nodes: registry.Nodes{n}},
				}
			}))
		}, ttl)
	})
}

func (m *mdnsRegistry) String() string { return Name }
