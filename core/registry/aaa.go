// Package registry 提供服务注册与撤销的抽象接口及 lifecycle 集成。
//
// 支持多种后端驱动（mdns 等），通过 Config.Driver 选择。
// 服务启动后自动注册到注册中心并定期续约，停止前自动撤销。
package registry

import (
	"context"

	"github.com/pubgo/lava/v2/core/service"
)

// Registry The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, mdns, ...}
type Registry interface {
	String() string
	Register(context.Context, *service.Service, ...RegOpt) error
	Deregister(context.Context, *service.Service, ...DeregOpt) error
}

type (
	Opt      func(*Opts)
	RegOpt   func(*RegOpts)
	DeregOpt func(*DeregOpts)
)
