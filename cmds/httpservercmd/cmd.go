package httpservercmd

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/funk/errors"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/supervisor"
	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/servers/https"
)

func New(di dix.Container) *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: cmdutil.UsageDesc("%s http service", version.Project()),
		Action: func(ctx context.Context, command *cli.Command) error {
			srv := dix.InjectMust(di, https.New())
			return errors.WrapCaller(supervisor.Run(ctx, srv))
		},
	}
}
