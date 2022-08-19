// Package registry is an interface for service discovery
package registry

import (
	"github.com/pubgo/funk/result"
)

// Registry The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, mdns, ...}
type Registry interface {
	String() string
	Register(*Service, ...RegOpt) error
	Deregister(*Service, ...DeregOpt) error
	Watch(string, ...WatchOpt) result.Result[Watcher]
	ListService(...ListOpt) result.List[*Service]
	GetService(string, ...GetOpt) result.List[*Service]
}

type Opt func(*Opts)
type RegOpt func(*RegOpts)
type WatchOpt func(*WatchOpts)
type DeregOpt func(*DeregOpts)
type GetOpt func(*GetOpts)
type ListOpt func(*ListOpts)
type Loader struct{}
