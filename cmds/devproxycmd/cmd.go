package devproxycmd

import (
	"context"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/redant"
)

func New(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "devproxy",
		Short: "Local development proxy tool",
		Long:  "Local development proxy tool with DNS and HTTP routing",
		Middleware: func(next redant.HandlerFunc) redant.HandlerFunc {
			return func(ctx context.Context, inv *redant.Invocation) error {
				assert.Must(loadConfig())

				return next(ctx, inv)
			}
		},
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			return StartDevProxy(ctx)
		},
		Children: []*redant.Command{
			{
				Use:   "start",
				Short: "Start devproxy server",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return StartDevProxy(ctx)
				},
			},
			{
				Use:   "install",
				Short: "Install system integration",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return InstallSystemIntegration()
				},
			},
			{
				Use:   "uninstall",
				Short: "Uninstall system integration",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return UninstallSystemIntegration()
				},
			},
			{
				Use:   "routes",
				Short: "Show current routes",
				Handler: func(ctx context.Context, i *redant.Invocation) error {
					return ShowRoutes()
				},
			},
		},
	}
}
