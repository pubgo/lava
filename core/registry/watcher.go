package registry

import (
	"github.com/pubgo/lava/core/service"
	"time"

	"github.com/pubgo/funk/result"

	"github.com/pubgo/lava/pkg/proto/event/v1"
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

// Event is registry event
type Event struct {
	// Id is registry id
	Id string
	// Type defines type of event
	Type eventpbv1.EventType
	// Timestamp is event timestamp
	Timestamp time.Time
	// Service is registry service
	Service *service.Service
}
