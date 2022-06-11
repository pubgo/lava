package service

import (
	"fmt"
	"os"
	"sort"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmd/cmds/healthcmd"
	"github.com/pubgo/lava/cmd/cmds/vercmd"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/version"
)

func Run(services ...Command) {
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
		cliApp.Commands = append(cliApp.Commands, cmd)
	}

	sort.Sort(cli.FlagsByName(cliApp.Flags))
	sort.Sort(cli.CommandsByName(cliApp.Commands))
	xerror.Panic(cliApp.Run(os.Args))
}
