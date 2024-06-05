package httpservercmd

import (
	"context"
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/servers/https"
	"github.com/urfave/cli/v3"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: cmdutil.UsageDesc("%s http service", version.Project()),
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()
			srv := dix.Inject(di, https.New())
			srv.Run()
			return nil
		},
	}
}
