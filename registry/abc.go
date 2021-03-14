// Package registry is an interface for service discovery
package registry

type Factory func(map[string]interface{}) (Registry, error)

// The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Register(*Service, ...RegOpt) error
	Deregister(*Service, ...DeRegOpt) error
	GetService(string, ...GetOpt) ([]*Service, error)
	ListServices(...ListOpt) ([]*Service, error)
	Watch(string, ...WatchOpt) (Watcher, error)
	String() string
}

type Opt func(*Opts)
type RegOpt func(*RegOpts)
type WatchOpt func(*WatchOpts)
type DeRegOpt func(*DeRegOpts)
type GetOpt func(*GetOpts)
type ListOpt func(*ListOpts)
