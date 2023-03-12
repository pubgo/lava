// Package mdns is a multicast dns registry
package mdns

import (
	"context"

	"github.com/grandcat/zeroconf"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/merge"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/typex"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/core/service"
)

const (
	zeroconfService  = "_lava._tcp"
	zeroconfDomain   = "local."
	zeroconfInstance = "lava"
)

func New(conf *registry.Config, log log.Logger) registry.Registry {
	if conf.Driver != Name {
		return nil
	}

	var cfg Cfg
	merge.MapStruct(&cfg, conf.DriverCfg).Unwrap()

	resolver, err := zeroconf.NewResolver()
	assert.MustF(err, "Failed to initialize zeroconf resolver")
	return &mdnsRegistry{resolver: resolver, cfg: cfg, log: log.WithName(registry.Name).WithName(Name)}
}

type serverNode struct {
	srv  *zeroconf.Server
	name string
	id   string
}

var _ registry.Registry = (*mdnsRegistry)(nil)

type mdnsRegistry struct {
	cfg      Cfg
	services typex.SyncMap
	resolver *zeroconf.Resolver
	log      log.Logger
}

func (m *mdnsRegistry) Close() {
}

func (m *mdnsRegistry) Init() {
}

func (m *mdnsRegistry) Register(ctx context.Context, service *service.Service, optList ...registry.RegOpt) (gErr error) {
	defer recovery.Recovery(func(err error) {
		gErr = errors.WrapKV(err, "service", service)
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

func (m *mdnsRegistry) Deregister(ctx context.Context, service *service.Service, opt ...registry.DeregOpt) (gErr error) {
	defer recovery.Recovery(func(err error) {
		gErr = errors.WrapKV(err, "service", service)
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

func (m *mdnsRegistry) String() string { return Name }
