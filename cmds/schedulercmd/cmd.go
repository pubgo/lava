package schedulercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/scheduler"
	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/pkg/cmdutil"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "scheduler",
		Usage: cmdutil.UsageDesc("grpc service %s(%s)", version.Project(), version.Version()),
		Action: func(ctx context.Context, command *cli.Command) error {
			srv := dix.Inject(di, new(struct {
				Scheduler *scheduler.Scheduler
			}))
			return errors.WrapCaller(supervisor.Run(ctx, srv.Scheduler))
		},
	}
}
