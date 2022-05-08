package service

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmd/cmds/vercmd"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/version"
)

type Command interface {
	Command() *cli.Command
}

func Run(services ...Command) {
	defer xerror.RespExit()

	xerror.Assert(len(services) == 0, "[services] is zero")

	for _, srv := range services {
		xerror.Assert(srv == nil, "[srv] is nil")
	}

	var app = &cli.App{
		Name:     runtime.Domain,
		Usage:    fmt.Sprintf("%s services", runtime.Domain),
		Version:  version.Version,
		Commands: []*cli.Command{vercmd.Cmd()},
	}

	for i := range services {
		srv := services[i]
		cmd := srv.Command()
		cmd.Before = func(context *cli.Context) error {
			runtime.Project = strings.TrimSpace(strings.Split(cmd.Name, " ")[0])

			mode := env.Get("lava_mode", "app_mode")
			if mode != "" {
				var i, err = strconv.Atoi(mode)
				xerror.Panic(err)

				runtime.Mode = runtime.RunMode(i)
				xerror.Assert(runtime.Mode.String() == "", "unknown mode, mode=%s", mode)
			}
			return nil
		}

		// 检查项目Command是否注册
		xerror.Assert(app.Command(cmd.Name) != nil, "command(%s) already exists", cmd.Name)
		app.Commands = append(app.Commands, cmd)
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	xerror.Panic(app.Run(os.Args))
}
