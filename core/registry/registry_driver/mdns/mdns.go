// Package mdns is a multicast dns registry
package mdns

import (
	"context"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/pubgo/x/try"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/pkg/typex"
)

const (
	zeroconfService  = "_lava._tcp"
	zeroconfDomain   = "local."
	zeroconfInstance = "lava"
)

func New(cfg Cfg) (registry.Registry, error) {
	resolver, err := zeroconf.NewResolver()
	xerror.Panic(err, "Failed to initialize zeroconf resolver")
	return &mdnsRegistry{resolver: resolver, cfg: cfg}, nil
}

var _ registry.Registry = (*mdnsRegistry)(nil)

type mdnsRegistry struct {
	cfg      Cfg
	services typex.SMap
	resolver *zeroconf.Resolver
}

func (m *mdnsRegistry) Close() {
}

func (m *mdnsRegistry) Init() {
}

func (m *mdnsRegistry) RegLoop(f func() *registry.Service, opt ...registry.RegOpt) error {
	return m.Register(f(), opt...)
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

func (m *mdnsRegistry) Deregister(service *registry.Service, opt ...registry.DeregOpt) (err error) {
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
	return services, xerror.Try(func() {
		entries := make(chan *zeroconf.ServiceEntry)
		syncx.GoSafe(func() {
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

func (m *mdnsRegistry) ListService(opts ...registry.ListOpt) (services []*registry.Service, _ error) {
	return services, nil
}

func (m *mdnsRegistry) Watch(service string, opt ...registry.WatchOpt) (w registry.Watcher, err error) {
	return w, try.Try(func() { w = newWatcher(m, service, opt...) })
}

func (m *mdnsRegistry) String() string { return Name }
