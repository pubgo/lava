// Package mdns is a multicast dns registry
package mdns

import (
	"context"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/syncx"
	"github.com/pubgo/funk/typex"
	"github.com/pubgo/funk/xerr"

	"github.com/pubgo/funk/result"
	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/logging/logutil"
)

const (
	zeroconfService  = "_lava._tcp"
	zeroconfDomain   = "local."
	zeroconfInstance = "lava"
)

func New(cfg Cfg, log *logging.Logger) registry.Registry {
	resolver, err := zeroconf.NewResolver()
	assert.MustF(err, "Failed to initialize zeroconf resolver")
	return &mdnsRegistry{resolver: resolver, cfg: cfg, log: logging.ModuleLog(log, logutil.Names(registry.Name, Name))}
}

type serverNode struct {
	srv  *zeroconf.Server
	name string
	id   string
}

var _ registry.Registry = (*mdnsRegistry)(nil)

type mdnsRegistry struct {
	cfg      Cfg
	services typex.SMap
	resolver *zeroconf.Resolver
	log      *logging.ModuleLogger
}

func (m *mdnsRegistry) Close() {
}

func (m *mdnsRegistry) Init() {
}

func (m *mdnsRegistry) Register(service *registry.Service, optList ...registry.RegOpt) (gErr result.Error) {
	defer recovery.Recovery(func(err xerr.XErr) {
		gErr = result.WithErr(err).WrapF("service=>%#v", service)
	})

	assert.If(service == nil, "[service] should not be nil")
	assert.If(len(service.Nodes) == 0, "[service] nodes should not be zero")

	node := service.Nodes[0]

	// 已经存在
	if m.services.Has(node.Id) {
		return
	}

	server, err := zeroconf.Register(node.Id, service.Name, zeroconfDomain, node.GetPort(), []string{node.Id}, nil)
	assert.MustF(err, "[mdns] service %s register error", service.Name)

	var opts registry.RegOpts
	for i := range optList {
		optList[i](&opts)
	}

	m.services.Set(node.Id, &serverNode{
		srv:  server,
		id:   node.Id,
		name: service.Name,
	})
	return
}

func (m *mdnsRegistry) Deregister(service *registry.Service, opt ...registry.DeregOpt) (gErr result.Error) {
	defer recovery.Recovery(func(err xerr.XErr) {
		gErr = result.WithErr(err.WrapF("service=>%#v", service))
	})

	assert.If(service == nil, "[service] should not be nil")
	assert.If(len(service.Nodes) == 0, "[service] nodes should not be zero")

	node := service.Nodes[0]
	var val, ok = m.services.LoadAndDelete(node.Id)
	if !ok || val == nil {
		return
	}

	val.(*serverNode).srv.Shutdown()
	return
}

func (m *mdnsRegistry) GetService(name string, opts ...registry.GetOpt) result.List[*registry.Service] {
	entries := make(chan *zeroconf.ServiceEntry)
	services := syncx.Yield(func(yield func(*registry.Service)) result.Error {
		for s := range entries {
			yield(&registry.Service{
				Name: s.Service,
				Nodes: registry.Nodes{{
					Id:      s.Instance,
					Port:    s.Port,
					Address: fmt.Sprintf("%s:%d", s.AddrIPv4[0].String(), s.Port),
				}},
			})
		}
		return result.NilErr()
	})

	var gOpts registry.GetOpts
	for i := range opts {
		opts[i](&gOpts)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	assert.MustF(m.resolver.Browse(ctx, name, zeroconfDomain, entries), "Failed to Lookup Service %s", name)
	<-ctx.Done()
	return services.ToList()
}

func (m *mdnsRegistry) ListService(opts ...registry.ListOpt) result.List[*registry.Service] {
	var services result.List[*registry.Service]
	m.services.Range(func(key, value interface{}) bool {
		services = append(services, m.GetService(key.(string))...)
		return true
	})
	return services
}

func (m *mdnsRegistry) Watch(service string, opt ...registry.WatchOpt) result.Result[registry.Watcher] {
	return newWatcher(m, service, opt...)
}

func (m *mdnsRegistry) String() string { return Name }
