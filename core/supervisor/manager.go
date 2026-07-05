package supervisor

import (
	"context"
	"sync"

	"github.com/pubgo/funk/v2/log"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/running"
)

func Default(lc lifecycle.Getter) *Manager {
	return NewManager(running.Project(), lc)
}

func NewManager(name string, lc lifecycle.Getter) *Manager {
	return &Manager{
		name:     name,
		lc:       lc,
		services: make(map[string]*serviceRunner),
		logger:   log.GetLogger(name),
	}
}

// Manager supervises lava services: registration, lifecycle hooks, and restart policy.
type Manager struct {
	name     string
	lc       lifecycle.Getter
	logger   log.Logger
	mu       sync.RWMutex
	services map[string]*serviceRunner

	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
}
