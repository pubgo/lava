package schedulercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/pkg/cmdutil"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "scheduler",
		Usage: cmdutil.UsageDesc("crontab scheduler service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			_ = dix.Inject(di, new(struct {
				Scheduler *scheduler.Scheduler
			}))

			return supervisor.Default().Run()
		},
	}
}
