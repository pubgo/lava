package grpcservercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/servers/tasks"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "task",
		Usage: cmdutil.UsageDesc("async task service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			di.Provide(tasks.New)
			params := dix.Inject(di, new(struct {
				LC       lifecycle.Getter
				Services []supervisor.Service
			}))

			manager := supervisor.Default(params.LC)
			for _, svc := range params.Services {
				assert.Exit(manager.Add(svc))
			}

			return manager.Run()
		},
	}
}
