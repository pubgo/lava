package registry

import (
	"github.com/pubgo/lava/types"

	"time"
)

// Watcher is an interface that returns updates
// about services within the registry.
type Watcher interface {
	// Next is a blocking call
	Next() (*Result, error)
	Stop() error
}

// Result is returned by a call to Next on
// the watcher. Actions can be create, update, delete
type Result struct {
	Action  types.EventType
	Service *Service
}

// Event is registry event
type Event struct {
	// Id is registry id
	Id string
	// Type defines type of event
	Type types.EventType
	// Timestamp is event timestamp
	Timestamp time.Time
	// Service is registry service
	Service *Service
}
