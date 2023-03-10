// Package registry is an interface for service discovery
package registry

import "github.com/pubgo/lava/core/service"

// Registry The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, mdns, ...}
type Registry interface {
	String() string
	Register(*service.Service, ...RegOpt) error
	Deregister(*service.Service, ...DeregOpt) error
}

type Opt func(*Opts)
type RegOpt func(*RegOpts)
type DeregOpt func(*DeregOpts)
type Loader struct{}
