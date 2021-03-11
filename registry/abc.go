// Package registry is an interface for service discovery
package registry

type Factory func(map[string]interface{}) (Registry, error)

// The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Register(*Service, ...RegisterOption) error
	Deregister(*Service, ...DeregisterOption) error
	GetService(string, ...GetOption) ([]*Service, error)
	ListServices(...ListOption) ([]*Service, error)
	Watch(...WatchOption) (Watcher, error)
	String() string
}

type Option func(*Options)
type RegisterOption func(*RegisterOptions)
type WatchOption func(*WatchOptions)
type DeregisterOption func(*DeregisterOptions)
type GetOption func(*GetOptions)
type ListOption func(*ListOptions)
