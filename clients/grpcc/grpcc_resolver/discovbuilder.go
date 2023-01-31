package grpcc_resolver

import (
	"context"
	"sync"

	"github.com/kr/pretty"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/result"
	"google.golang.org/grpc/resolver"

	"github.com/pubgo/lava/core/registry"
	"github.com/pubgo/lava/pkg/proto/event/v1"
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
	defer recovery.Recovery(func(err xerr.XErr) {
		gErr = err
		pretty.Println(target.URL.String())
	})

	logs.S().Infof("discovBuilder Build, target=>%#v", target)

	// 直接通过全局变量[registry.Default]获取注册中心, 然后进行判断
	var r = registry.Default()
	assert.If(r == nil, "registry is nil")

	var srv = target.URL.Host

	// target.Endpoint是服务的名字, 是项目启动的时候注册中心中注册的项目名字
	// GetService根据服务名字获取注册中心该项目所有服务
	services := r.GetService(srv).ToResult()

	// 启动后，更新服务地址
	d.updateService(services.Expect("registry GetService error")...)

	var address = d.getAddrList(srv)
	assert.If(len(address) == 0, "service none available")

	logs.S().Infof("discovBuilder Addrs %#v", address)
	assert.MustF(cc.UpdateState(newState(address)), "update resolver address: %v", address)

	w := r.Watch(srv).Unwrap(func(err result.Error) result.Error {
		return err.WrapF("target.Endpoint: %s", srv)
	})

	return &baseResolver{
		cancel: syncx.GoCtx(func(ctx context.Context) (gErr result.Error) {
			defer func() { w.Stop() }()

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
						logs.WithErr(res.Err()).Error("error")
						continue
					}

					// 注册中心删除服务
					if res.Unwrap().Action == eventpbv1.EventType_DELETE {
						d.delService(res.Unwrap().Service)
					} else {
						d.updateService(res.Unwrap().Service)
					}

					xtry.TryErr(func() result.Error {
						var addrList = d.getAddrList(srv)
						return result.WithErr(cc.UpdateState(newState(addrList)))
					}).Do(func(err result.Error) {
						logs.WithErr(err.Unwrap()).Error("update resolver address error")
					})
				}
			}
		}),
		builder: DiscovScheme,
	}, nil
}
