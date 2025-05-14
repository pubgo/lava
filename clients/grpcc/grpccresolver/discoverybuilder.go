package grpccresolver

import (
	"context"
	"strings"
	"sync"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/pretty"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/try"
	"google.golang.org/grpc/resolver"

	"github.com/pubgo/lava/core/discovery"
	"github.com/pubgo/lava/core/service"
	"github.com/pubgo/lava/internal/logutil"
	"github.com/pubgo/lava/pkg/proto/lavapbv1"
)

func NewDiscoveryBuilder(disco discovery.Discovery) resolver.Builder {
	return &discoveryBuilder{
		log:   logs.WithName(DiscoveryScheme),
		disco: disco,
	}
}

var _ resolver.Builder = (*discoveryBuilder)(nil)

type discoveryBuilder struct {
	// getServiceUniqueId -> *resolver.Address
	services sync.Map
	disco    discovery.Discovery
	log      log.Logger
}

func (d *discoveryBuilder) Scheme() string { return DiscoveryScheme }

// 删除服务
func (d *discoveryBuilder) delService(services ...*service.Service) {
	for i := range services {
		for _, n := range services[i].Nodes {
			// 删除服务信息
			for j := 0; j < Replica; j++ {
				d.services.Delete(getServiceUniqueId(n.Id, j))
			}
		}
	}
}

// 更新服务
func (d *discoveryBuilder) updateService(services ...*service.Service) {
	for i := range services {
		for _, n := range services[i].Nodes {
			// 更新服务信息
			for j := 0; j < Replica; j++ {
				addr := n.Address
				res := newAddr(addr, services[i].Name)
				val, ok := d.services.LoadOrStore(getServiceUniqueId(n.Id, j), &res)
				if ok {
					val.(*resolver.Address).Addr = addr
					val.(*resolver.Address).ServerName = services[i].Name
				}
			}
		}
	}
}

// 获取服务地址
func (d *discoveryBuilder) getAddrList(name string) []resolver.Address {
	var addrList []resolver.Address
	d.services.Range(func(_, value interface{}) bool {
		addr := *value.(*resolver.Address)
		if addr.ServerName == name {
			addrList = append(addrList, *value.(*resolver.Address))
		}
		return true
	})
	return addrList
}

// Build discovery://service_name:50051
func (d *discoveryBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, gErr error) {
	defer recovery.Recovery(func(err error) {
		gErr = err
		pretty.Println(target.URL.String())
	})

	// 服务发现
	logs.Info().Msgf("discovery builder, target=>%#v", target)

	assert.If(d.disco == nil, "registry is nil")
	srv := strings.SplitN(target.URL.Host, ":", 2)[0]

	// target.Endpoint是服务的名字, 是项目启动的时候注册中心中注册的项目名字
	// GetService根据服务名字获取注册中心该项目所有服务
	services := d.disco.GetService(context.Background(), srv).Unwrap(func(err error) error {
		return errors.Wrapf(err, "failed to GetService, srv=%s", srv)
	})

	// 启动后，更新服务地址
	d.updateService(services...)

	address := d.getAddrList(srv)
	assert.If(len(address) == 0, "service none available")

	logs.Info().Msgf("discovery builder UpdateState, address=%v", address)
	assert.MustF(cc.UpdateState(newState(address)), "update resolver address: %v", address)

	w := d.disco.Watch(context.Background(), srv).Unwrap(func(err error) error {
		return errors.Wrapf(err, "target.Endpoint: %s", srv)
	})

	return &baseResolver{
		serviceName: srv,
		builder:     DiscoveryScheme,
		cancel: async.GoCtx(func(ctx context.Context) (gErr error) {
			defer logutil.HandleClose(logs, w.Stop)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					res := w.Next()
					if res.IsErr() {
						if errors.Is(res.Err(), discovery.ErrWatcherStopped) {
							return
						}

						d.log.Err(res.Err(), ctx).Msg("failed to get service watcher event")

						if errors.Is(res.Err(), discovery.ErrTimeout) {
							continue
						}

						continue
					}

					// 注册中心删除服务
					if res.Unwrap().Action == lavapbv1.EventType_DELETE {
						d.delService(res.Unwrap().Service)
					} else {
						d.updateService(res.Unwrap().Service)
					}

					logutil.ErrRecord(logs, try.Try(func() error {
						addrList := d.getAddrList(srv)
						return cc.UpdateState(newState(addrList))
					}), func(evt *log.Event) string {
						return "failed to update resolver address"
					})
				}
			}
		}),
	}, nil
}
