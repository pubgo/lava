package resolver

import (
	"fmt"
	"sync"

	"github.com/pubgo/lug/registry"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"google.golang.org/grpc/resolver"
)

type discovBuilder struct {
	// getServiceUniqueId -> *resolver.Address
	services sync.Map
}

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
				// 如果port不存在, 那么addr中包含port
				if n.Port > 0 {
					addr = fmt.Sprintf("%s:%d", n.Address, n.Port)
				}

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
func (d *discovBuilder) getAddrs() []resolver.Address {
	var addrs []resolver.Address
	d.services.Range(func(_, value interface{}) bool {
		addrs = append(addrs, *value.(*resolver.Address))
		return true
	})
	return reshuffle(addrs)
}

// discovBuilder discov://wpt.etcd/service_name
func (d *discovBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (resolver.Resolver, error) {
	// target.Authority得到注册中心的地址
	// 当然也可以直接通过全局变量[registry.Default]获取注册中心, 然后进行判断
	var r = registry.Default
	xerror.Assert(r == nil, "registry %s not found", target.Authority)

	// target.Endpoint是服务的名字, 是项目启动的时候注册中心中注册的项目名字
	// GetService根据服务名字获取注册中心该项目所有服务
	services, err := r.GetService(target.Endpoint)
	if err != nil {
		return nil, xerror.Wrap(err, "registry GetService error\n")
	}

	// 启动后，更新服务地址
	d.updateService(services...)

	var addrs = d.getAddrs()

	if len(addrs) == 0 {
		return nil, fmt.Errorf("service none available")
	}

	cc.UpdateState(resolver.State{Addresses: addrs})

	w, err := r.Watch(target.Endpoint)
	if err != nil {
		return nil, xerror.WrapF(err, "target.Endpoint:%s\n", target.Endpoint)
	}

	go func() {
		defer w.Stop()
		for {
			res, err := w.Next()
			if err == registry.ErrWatcherStopped {
				break
			}

			if err != nil {
				xlog.Error(err.Error())
				continue
			}

			// 注册中心删除服务
			if res.Action == "delete" {
				d.delService(res.Service)
			} else {
				d.updateService(res.Service)
			}

			cc.UpdateState(resolver.State{Addresses: d.getAddrs()})
		}
	}()

	return &baseResolver{cc: cc, r: w}, nil
}

func (d *discovBuilder) Scheme() string { return DiscovScheme }
