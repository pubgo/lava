package schedulercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/pkg/cmdutil"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "scheduler",
		Usage: cmdutil.UsageDesc("crontab scheduler service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			di.Provide(scheduler.NewService)
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
