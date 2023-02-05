package grpcc_resolver

import (
	"context"
	"sync"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/log/logutil"
	"github.com/pubgo/funk/pretty"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/try"
	"github.com/pubgo/lava/core/registry"
	eventpbv1 "github.com/pubgo/lava/pkg/proto/event/v1"
	"google.golang.org/grpc/resolver"
)

var _ resolver.Builder = (*discovBuilder)(nil)

type discovBuilder struct {
	// getServiceUniqueId -> *resolver.Address
	services sync.Map
}

func (d *discovBuilder) Scheme() string { return DiscovScheme }

// 删除服务
func (d *discovBuilder) delService(services ...*registry.Service) {
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
func (d *discovBuilder) updateService(services ...*registry.Service) {
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
func (d *discovBuilder) getAddrList(name string) []resolver.Address {
	var addrList []resolver.Address
	d.services.Range(func(_, value interface{}) bool {
		var addr = *value.(*resolver.Address)
		if addr.ServerName == name {
			addrList = append(addrList, *value.(*resolver.Address))
		}
		return true
	})
	return addrList
}

// Build discov://service_name
func (d *discovBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, gErr error) {
	defer recovery.Recovery(func(err error) {
		gErr = err
		pretty.Println(target.URL.String())
	})

	// 服务发现
	logs.Info().Msgf("discovery builder, target=>%#v", target)

	// 直接通过全局变量[registry.Default]获取注册中心, 然后进行判断
	var r = registry.Default()
	assert.If(r == nil, "registry is nil")

	var srv = target.URL.Host

	// target.Endpoint是服务的名字, 是项目启动的时候注册中心中注册的项目名字
	// GetService根据服务名字获取注册中心该项目所有服务
	services := r.GetService(srv).Unwrap(func(err error) error {
		return errors.Wrapf(err, "failed to GetService, srv=%s", srv)
	})

	// 启动后，更新服务地址
	d.updateService(services...)

	var address = d.getAddrList(srv)
	assert.If(len(address) == 0, "service none available")

	logs.Info().Msgf("discovery builder UpdateState, address=%v", address)
	assert.MustF(cc.UpdateState(newState(address)), "update resolver address: %v", address)

	w := r.Watch(srv).Unwrap(func(err error) error {
		return errors.Wrapf(err, "target.Endpoint: %s", srv)
	})

	return &baseResolver{
		cancel: async.GoCtx(func(ctx context.Context) (gErr error) {
			defer logutil.HandleClose(logs, w.Stop)

			for {
				select {
				case <-ctx.Done():
					return
				default:
					res := w.Next()
					if res.Err() == registry.ErrWatcherStopped {
						return
					}

					if res.IsErr() {
						logutil.ErrRecord(logs, res.Err(), func(evt *log.Event) string {
							return "failed to get watcher service event"
						})
						continue
					}

					// 注册中心删除服务
					if res.Unwrap().Action == eventpbv1.EventType_DELETE {
						d.delService(res.Unwrap().Service)
					} else {
						d.updateService(res.Unwrap().Service)
					}

					logutil.ErrRecord(logs, try.Try(func() error {
						var addrList = d.getAddrList(srv)
						return cc.UpdateState(newState(addrList))
					}), func(evt *log.Event) string {
						return "failed to update resolver address"
					})
				}
			}
		}),
		builder: DiscovScheme,
	}, nil
}
