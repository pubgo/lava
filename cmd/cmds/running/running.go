package running

import (
	"fmt"
	"os"
	"sort"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmd/cmds/healthcmd"
	"github.com/pubgo/lava/cmd/cmds/migrate"
	"github.com/pubgo/lava/cmd/cmds/vercmd"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/core/signal"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
)

func Run(services ...service.Command) {
	defer xerror.RecoverAndExit()

	xerror.Assert(len(services) == 0, "[services] is zero")

	for _, srv := range services {
		xerror.Assert(srv == nil, "[srv] is nil")
	}

	var cliApp = &cli.App{
		Name:     runmode.Domain,
		Usage:    fmt.Sprintf("%s services", runmode.Domain),
		Version:  version.Version,
		Commands: []*cli.Command{vercmd.Cmd(), healthcmd.Cmd()},
	}

	for i := range services {
		srv := services[i]
		cmd := srv.Command()

		// 检查项目Command是否注册
		xerror.Assert(cliApp.Command(cmd.Name) != nil, "command(%s) already exists", cmd.Name)

		cmd.Action = func(ctx *cli.Context) error {
			defer xerror.RecoverAndExit()
			xerror.Panic(srv.Start())
			signal.Block()
			xerror.Panic(srv.Stop())
			return nil
		}

		cmd.Subcommands = append(cmd.Subcommands, migrate.Cmd())
		cliApp.Commands = append(cliApp.Commands, cmd)
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))
	xerror.Panic(cliApp.Run(os.Args))
}
