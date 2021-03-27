// Package mdns is a multicast dns registry
package mdns

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/internal/gutils"
	"github.com/pubgo/golug/registry"
	"github.com/pubgo/golug/types"
	"github.com/pubgo/x/abc"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
)

func init() {
	registry.Register(Name, func(m map[string]interface{}) (registry.Registry, error) {
		resolver, err := zeroconf.NewResolver()
		xerror.Panic(err, "Failed to initialize resolver")

		var r = &mdnsRegistry{resolver: resolver}

		xerror.Panic(gutils.Map(&r.cfg, m))
		return r, nil
	})
}

var _ registry.Registry = (*mdnsRegistry)(nil)
var _ registry.Watcher = (*mdnsRegistry)(nil)

type mdnsRegistry struct {
	watcher  chan *registry.Result
	mu       sync.Mutex
	cfg      Cfg
	resolver *zeroconf.Resolver
	services map[string]*zeroconf.Server
	cancel   *abc.Cancel
}

func (m *mdnsRegistry) Next() (*registry.Result, error) {
	result, ok := <-m.watcher

	if !ok {
		return nil, registry.ErrWatcherStopped
	}

	return result, nil
}

func (m *mdnsRegistry) Stop() {
	close(m.watcher)
	if m.cancel != nil {
		m.cancel.Cancel()
	}
}

func (m *mdnsRegistry) Register(service *registry.Service, opt ...registry.RegOpt) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return xutil.Try(func() {
		xerror.Assert(service == nil, "[service] should not be nil")
		xerror.Assert(len(service.Nodes) == 0, "service nodes should not be zero")

		node := service.Nodes[0]
		server, err := zeroconf.Register(
			node.Id,
			service.Name,
			"local",
			node.GetPort(),
			[]string{"register"},
			nil,
		)
		xerror.PanicF(err, "[mdns] service %s register error", service.Name)

		m.services[node.Id] = server

		var opts registry.RegOpts
		for i := range opt {
			opt[i](&opts)
		}

		if opts.TTL != 0 {
			server.TTL(uint32(opts.TTL.Seconds()))
		}
	})
}

func (m *mdnsRegistry) Deregister(service *registry.Service, opt ...registry.DeRegOpt) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	return xutil.Try(func() {
		xerror.Assert(service == nil, "[service] should not be nil")
		xerror.Assert(len(service.Nodes) == 0, "service nodes should not be zero")

		node := service.Nodes[0]
		m.services[node.Id].Shutdown()
	})
}

func (m *mdnsRegistry) GetService(s string, opts ...registry.GetOpt) ([]*registry.Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var services []*registry.Service
	return services, xutil.Try(func() {
		entries := make(chan *zeroconf.ServiceEntry)
		go func(results <-chan *zeroconf.ServiceEntry) {
			for s := range results {
				services = append(services, &registry.Service{
					Name: s.Service,
					Nodes: registry.NodeOf(&registry.Node{
						Id:      s.Instance,
						Port:    s.Port,
						Address: fmt.Sprintf("%s:%d", s.AddrIPv4[0].String(), s.Port),
					}),
				})
			}
		}(entries)

		var gOpts registry.GetOpts
		for i := range opts {
			opts[i](&gOpts)
		}

		if gOpts.Timeout == 0 {
			gOpts.Timeout = time.Second * 5
		}

		ctx, cancel := context.WithTimeout(context.Background(), gOpts.Timeout)
		defer cancel()

		xerror.Panic(m.resolver.Browse(ctx, s, config.Domain, entries), "Failed to Browse")
		<-ctx.Done()
	})
}

func (m *mdnsRegistry) ListServices(opt ...registry.ListOpt) ([]*registry.Service, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	return nil, errors.New("[mdns] ListServices not implemented")
}

func (m *mdnsRegistry) Watch(service string, opt ...registry.WatchOpt) (registry.Watcher, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var watcher = &mdnsRegistry{watcher: make(chan *registry.Result)}

	return watcher, xutil.Try(func() {
		xerror.Assert(service == "", "[service] should not be null")

		var allNodes types.SMap
		services, err := m.GetService(service)
		xerror.Panic(err)
		for i := range services {
			for _, n := range services[i].Nodes {
				allNodes.Set(n.Id, n)
			}
		}

		watcher.cancel = fx.GoLoop(func(ctx context.Context) {
			var nodes types.SMap

			select {
			case <-time.Tick(m.cfg.TTL):
				entries := make(chan *zeroconf.ServiceEntry)
				go func(results <-chan *zeroconf.ServiceEntry) {
					for s := range results {
						nodes.Set(s.Instance, &registry.Node{
							Id:      s.Instance,
							Port:    s.Port,
							Address: fmt.Sprintf("%s:%d", s.AddrIPv4[0].String(), s.Port),
						})
					}
				}(entries)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
				defer cancel()
				xerror.Panic(m.resolver.Browse(ctx, service, "local", entries), "Failed to Browse")
				<-ctx.Done()

				xerror.Panic(nodes.Each(func(id string, n *registry.Node) {
					if allNodes.Has(id) {
						return
					}

					allNodes.Set(id, n)
					watcher.watcher <- &registry.Result{
						Action: registry.Update.String(),
						Service: &registry.Service{
							Name:  service,
							Nodes: registry.NodeOf(n),
						},
					}
				}))

				xerror.Panic(allNodes.Each(func(id string, n *registry.Node) {
					if nodes.Has(id) {
						return
					}

					allNodes.Delete(id)
					watcher.watcher <- &registry.Result{
						Action: registry.Delete.String(),
						Service: &registry.Service{
							Name:  service,
							Nodes: registry.NodeOf(n),
						},
					}
				}))
			}
		})
	})
}

func (m *mdnsRegistry) String() string { return Name }
