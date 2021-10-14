package resolver

import (
	"context"
	"fmt"
	"sync"

	"github.com/pubgo/xerror"
	"go.uber.org/zap"
	"google.golang.org/grpc/resolver"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugins/registry"
	"github.com/pubgo/lava/types"
)

var logs *zap.Logger

func init() {
	logs = logger.On(func(log *zap.Logger) { logs = log.Named("balancer.resolver") })
}

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
				// 如果port不存在, 那么addr中包含port
				//if !strings.Contains(n.Address, ":") {
				addr = fmt.Sprintf("%s:%d", "localhost", n.Port)
				//}

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
func (d *discovBuilder) Build(target resolver.Target, cc resolver.ClientConn, opts resolver.BuildOptions) (_ resolver.Resolver, err error) {
	defer xerror.RespErr(&err)

	logs.Sugar().Debugf("discovBuilder Build %#v", target)

	// 直接通过全局变量[registry.Default]获取注册中心, 然后进行判断
	var r = registry.Default()
	xerror.Assert(r == nil, "registry is nil")

	// target.Endpoint是服务的名字, 是项目启动的时候注册中心中注册的项目名字
	// GetService根据服务名字获取注册中心该项目所有服务
	services, err := r.GetService(target.Endpoint)
	xerror.Panic(err, "registry GetService error")

	// 启动后，更新服务地址
	d.updateService(services...)

	var addrs = d.getAddrList(target.Endpoint)
	xerror.Assert(len(addrs) == 0, "service none available")

	logs.Sugar().Infof("discovBuilder Addrs %#v", addrs)
	xerror.PanicF(cc.UpdateState(newState(addrs)), "update resolver address: %v", addrs)

	w, err := r.Watch(target.Endpoint)
	xerror.PanicF(err, "target.Endpoint: %s", target.Endpoint)

	cancel := syncx.GoCtx(func(ctx context.Context) {
		defer func() { xerror.Panic(w.Stop()) }()

		for {
			select {
			case <-ctx.Done():
				return
			default:
				res, err := w.Next()
				if err == registry.ErrWatcherStopped {
					return
				}

				if err != nil {
					logs.Error("error", zap.Any("err", err))
					continue
				}

				// 注册中心删除服务
				if res.Action == types.EventType_DELETE {
					d.delService(res.Service)
				} else {
					d.updateService(res.Service)
				}

				xerror.TryCatch(func() {
					var addrs = d.getAddrList(target.Endpoint)
					xerror.PanicF(cc.UpdateState(newState(addrs)), "update resolver address: %v", addrs)
				}, func(err error) {
					logz.WithErr("balancer.resolver", err).Error("update state error")
				})
			}
		}
	})

	return &baseResolver{cc: cc, r: w, cancel: cancel, builder: DiscovScheme}, nil
}
