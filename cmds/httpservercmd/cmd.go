package httpservercmd

import (
	"fmt"

	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/servers/https"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: fmt.Sprintf("%s http service", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			srv := di.Inject(https.New())
			srv.Run()
			return nil
		},
	}
}
