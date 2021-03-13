// Package mdns is a multicast dns registry
package mdns

import (
	"context"
	"sync"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/golug/registry"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/xutil"
	"github.com/pubgo/xerror"
)

const Name = "mdns"

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
}

func (m *mdnsRegistry) Next() (*registry.Result, error) {
	result, ok := <-m.watcher

	if !ok {
		return nil, registry.ErrWatcherStopped
	}

	return result, nil
}

func (m *mdnsRegistry) Stop() {
	panic("implement me")
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
		xerror.Panic(err, "[mdns] service %s register error", service.Name)

		m.services[node.Id] = server

		var opts registry.RegOpts
		for i := range opt {
			opt[i](&opts)
		}

		if opts.TTL != 0 {
			server.TTL(uint32(opts.TTL.Seconds()))
		}

		if len(m.cfg.Text) > 0 {
			server.SetText(m.cfg.Text)
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

func (m *mdnsRegistry) GetService(s string, opt ...registry.GetOpt) ([]*registry.Service, error) {
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
						Id:   s.Instance,
						Port: s.Port,
					}),
				})
			}
		}(entries)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()

		xerror.Panic(m.resolver.Browse(ctx, s, config.Domain, entries), "Failed to Lookup")

		<-ctx.Done()
	})
}

func (m *mdnsRegistry) ListServices(opt ...registry.ListOpt) ([]*registry.Service, error) {
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
						Id:   s.Instance,
						Port: s.Port,
					}),
				})
			}
		}(entries)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
		defer cancel()

		xerror.Panic(m.resolver.Browse(ctx, "", config.Domain, entries), "Failed to Lookup")

		<-ctx.Done()
	})
}

func (m *mdnsRegistry) Watch(opt ...registry.WatchOpt) (registry.Watcher, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	var watcher = &mdnsRegistry{watcher: make(chan *registry.Result)}

	return watcher, xutil.Try(func() {
		_ = fx.GoLoop(func(ctx context.Context) {
			select {
			case <-time.Tick(m.cfg.TTL):
				entries := make(chan *zeroconf.ServiceEntry)
				go func(results <-chan *zeroconf.ServiceEntry) {
					for s := range results {
						watcher.watcher <- &registry.Result{
							Action: "",
							Service: &registry.Service{
								Name: s.Service,
								Nodes: registry.NodeOf(&registry.Node{
									Id:   s.Instance,
									Port: s.Port,
								}),
							},
						}
					}
				}(entries)

				ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
				defer cancel()

				xerror.Panic(m.resolver.Browse(ctx, "", config.Domain, entries), "Failed to Lookup")

				<-ctx.Done()
			}
		})
	})
}

func (m *mdnsRegistry) String() string { return Name }
