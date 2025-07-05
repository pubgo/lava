package discovery

import (
	"context"
	"time"

	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/lava/core/service"
	"github.com/pubgo/lava/pkg/proto/lavapbv1"
)

type (
	WatchOpt func(*WatchOpts)
	GetOpt   func(*GetOpts)
)

type Discovery interface {
	String() string
	Watch(ctx context.Context, srv string, opts ...WatchOpt) result.Result[Watcher]
	GetService(ctx context.Context, srv string, opts ...GetOpt) result.Result[[]*service.Service]
}

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
	Action  lavapbv1.EventType
	Service *service.Service
}

type WatchOpts struct {
	Service string
}

type GetOpts struct {
	Timeout time.Duration
}
