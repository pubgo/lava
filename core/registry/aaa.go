// Package registry is an interface for service discovery
package registry

import (
	"context"

	"github.com/pubgo/lava/core/service"
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
