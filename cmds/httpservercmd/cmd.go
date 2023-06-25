package httpservercmd

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/pkg/cmds"
	"github.com/pubgo/lava/servers/https"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: cmds.UsageDesc("%s http service", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			srv := dix.Inject(di, https.New())
			srv.Run()
			return nil
		},
	}
}
