// Package registry is an interface for service discovery
package registry

// Registry The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, mdns, ...}
type Registry interface {
	Init()
	String() string
	Register(*Service, ...RegOpt) error
	RegLoop(func() *Service, ...RegOpt) error
	Deregister(*Service, ...DeregOpt) error
	Watch(string, ...WatchOpt) (Watcher, error)
	ListService(...ListOpt) ([]*Service, error)
	GetService(string, ...GetOpt) ([]*Service, error)
}

type Opt func(*Opts)
type RegOpt func(*RegOpts)
type WatchOpt func(*WatchOpts)
type DeregOpt func(*DeregOpts)
type GetOpt func(*GetOpts)
type ListOpt func(*ListOpts)
