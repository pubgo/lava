// Package bootstrap provides shared supervisor wiring for CLI commands.
package bootstrap

import (
	"context"

	"github.com/pubgo/funk/v2/assert"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/supervisor"
	supervisordebug "github.com/pubgo/lava/v2/core/supervisor/debug"
)

// Prepare returns a Manager with debug endpoints registered.
func Prepare(lc lifecycle.Getter) *supervisor.Manager {
	manager := supervisor.Default(lc)
	supervisordebug.Register(manager)
	return manager
}

// Run starts all services on a prepared manager.
func Run(ctx context.Context, manager *supervisor.Manager) error {
	return manager.Run(ctx)
}

// RunServices creates a manager, registers services, and runs until shutdown.
func RunServices(ctx context.Context, lc lifecycle.Getter, services []supervisor.Service) error {
	manager := Prepare(lc)
	for _, svc := range services {
		assert.Exit(manager.Add(svc))
	}
	return Run(ctx, manager)
}
