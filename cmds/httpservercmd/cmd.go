package httpservercmd

import (
	"context"

	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/core/supervisor"
	supervisordebug "github.com/pubgo/lava/v2/core/supervisor/debug"
	"github.com/pubgo/lava/v2/pkg/cliutil"
	"github.com/pubgo/lava/v2/servers/https"
)

func New(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "http",
		Short: cliutil.UsageDesc("%s http service", version.Project()),
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			di.Provide(https.New)
			di.Provide(supervisor.Default)
			params := dix.Inject(di, new(struct {
				Services   []supervisor.Service
				Supervisor *supervisor.Manager
			}))

			supervisordebug.Register(params.Supervisor)
			for _, svc := range params.Services {
				assert.Exit(params.Supervisor.Add(svc))
			}

			return params.Supervisor.Run(ctx)
		},
	}
}
