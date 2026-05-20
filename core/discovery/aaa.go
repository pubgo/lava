package discovery

import (
	"context"
	"time"

	"github.com/pubgo/funk/v2/result"

	"github.com/pubgo/lava/v2/core/service"
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
	Action  EventType
	Service *service.Service
}

type WatchOpts struct {
	Service string
}

type GetOpts struct {
	Timeout time.Duration
}

type EventType int32

const (
	EventType_UNKNOWN EventType = 0
	EventType_CREATE  EventType = 1
	EventType_UPDATE  EventType = 2
	EventType_DELETE  EventType = 3
)
