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
	Watch(srv string, opts ...WatchOpt) result.Result[Watcher]
	ListService(opts ...ListOpt) result.Result[[]*service.Service]
	GetService(srv string, opts ...GetOpt) result.Result[[]*service.Service]
}

type WatchOpt func(*WatchOpts)
type GetOpt func(*GetOpts)
type ListOpt func(*ListOpts)

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
	Context context.Context
}

type GetOpts struct {
	Timeout time.Duration
	Context context.Context
}

type ListOpts struct {
	Context context.Context
}

// WatchService Watch a service
func WatchService(name string) WatchOpt {
	return func(o *WatchOpts) {
		o.Service = name
	}
}

func WatchContext(ctx context.Context) WatchOpt {
	return func(o *WatchOpts) {
		o.Context = ctx
	}
}

func ListContext(ctx context.Context) ListOpt {
	return func(o *ListOpts) {
		o.Context = ctx
	}
}
