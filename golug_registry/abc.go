// Package registry is an interface for service discovery
package golug_registry

import (
	"context"
	"errors"
	"time"
)

// The registry provides an interface for service discovery
// and an abstraction over varying implementations
// {consul, etcd, zookeeper, ...}
type Registry interface {
	Init(cfg Cfg) error
	Options() Cfg
	Register(*Service, ...RegisterOption) error
	Deregister(*Service) error
	GetService(string) ([]*Service, error)
	ListServices() ([]*Service, error)
	Watch(...WatchOption) (Watcher, error)
	String() string
}

type RegisterOptions struct {
	TTL time.Duration
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

type WatchOptions struct {
	// Specify a service to watch
	// If blank, the watch is for all services
	Service string
	// Other options for implementations of the interface
	// can be stored in a context
	Context context.Context
}

type RegisterOption func(*RegisterOptions)

type WatchOption func(*WatchOptions)

// Not found error when GetService is called
var ErrNotFound = errors.New("not found")

// Watcher stopped error when watcher is stopped
var ErrWatcherStopped = errors.New("watcher stopped")
