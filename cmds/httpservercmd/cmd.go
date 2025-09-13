package httpservercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/v2/core/lifecycle"
	"github.com/pubgo/lava/v2/core/supervisor"
	"github.com/pubgo/lava/v2/pkg/cmdutil"
	"github.com/pubgo/lava/v2/servers/https"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: cmdutil.UsageDesc("%s http service", version.Project()),
		Action: func(ctx context.Context, command *cli.Command) error {
			di.Provide(https.New)
			params := dix.Inject(di, new(struct {
				LC       lifecycle.Getter
				Services []supervisor.Service
			}))

			manager := supervisor.Default(params.LC)
			for _, svc := range params.Services {
				assert.Exit(manager.Add(svc))
			}

			return manager.Run(ctx)
		},
	}
}
