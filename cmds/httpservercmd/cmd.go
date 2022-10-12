package httpservercmd

import (
	"fmt"

	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/servers/https"
	"github.com/pubgo/lava/version"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "http",
		Usage: fmt.Sprintf("%s http service", version.Project()),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			srv := di.Inject(https.New())
			srv.Start()
			signal.Wait()
			srv.Stop()
			return nil
		},
	}
}
