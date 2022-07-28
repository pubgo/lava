package running

import (
	"fmt"
	"os"
	"sort"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmd/cmds/healthcmd"
	"github.com/pubgo/lava/cmd/cmds/migratecmd"
	"github.com/pubgo/lava/cmd/cmds/vercmd"
	"github.com/pubgo/lava/core/flags"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func Run(srv service.Runtime, cmds ...*cli.Command) {
	defer recovery.Exit()

	var serveCmd = &cli.Command{
		Name: "serve",
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()
			srv.Start()
			signal.Wait()
			srv.Stop()
			return nil
		},
	}

	var app = &cli.App{
		Name:     version.Project(),
		Usage:    fmt.Sprintf("%s service", version.Project()),
		Version:  version.Version(),
		Flags:    flags.GetFlags(),
		Commands: append(cmds, serveCmd, vercmd.Cmd(), healthcmd.Cmd(), migratecmd.Cmd()),
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	assert.Must(app.Run(os.Args))
}
