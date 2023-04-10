package discovery

import (
	"context"
	"time"

	"github.com/pubgo/funk/result"
	"github.com/pubgo/lava/core/service"
	eventpbv1 "github.com/pubgo/lava/pkg/proto/event/v1"
	//	 https://github.com/prometheus/prometheus/tree/main/discovery
)

type Discovery interface {
	String() string
	Watch(ctx context.Context, srv string, opts ...WatchOpt) result.Result[Watcher]
	GetService(ctx context.Context, srv string, opts ...GetOpt) result.Result[[]*service.Service]
}

type (
	WatchOpt func(*WatchOpts)
	GetOpt   func(*GetOpts)
)

// Watcher is an interface that returns updates
// about services within the registry.
type Watcher interface {
	// Next is a blocking call
	Next() result.Result[*Result]
	Stop() error
}

// Result is returned by a call to Next on
// the watcher. Actions can be create, update, delete
type Result struct {
	Action  eventpbv1.EventType
	Service *service.Service
}

type WatchOpts struct {
	Service string
}

type GetOpts struct {
	Timeout time.Duration
}

// ServiceInstance is an instance of a service in a discovery system.
type ServiceInstance struct {
	// ID is the unique instance ID as registered.
	ID string `json:"id"`
	// Name is the service name as registered.
	Name string `json:"name"`
	// Version is the version of the compiled.
	Version string `json:"version"`
	// Metadata is the kv pair metadata associated with the service instance.
	Metadata map[string]string `json:"metadata"`
	// Endpoints are endpoint addresses of the service instance.
	// schema:
	//   http://127.0.0.1:8000?isSecure=false
	//   grpc://127.0.0.1:9000?isSecure=false
	Endpoints []string `json:"endpoints"`
}
