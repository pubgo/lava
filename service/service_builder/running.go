package service_builder

import (
	"fmt"
	"os"
	"sort"

	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/version"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
	_ "go.uber.org/fx"
)

func Run(services ...service.Service) {
	defer xerror.RespExit()

	xerror.Assert(len(services) == 0, "[services] is zero")

	for _, srv := range services {
		xerror.Assert(srv == nil, "[srv] is nil")
	}

	var app = &cli.App{
		Name:    runtime.Domain,
		Usage:   fmt.Sprintf("%s services", runtime.Domain),
		Version: version.Version,
	}

	for i := range services {
		srv := services[i]
		cmd := srv.Command()

		// 检查项目Command是否注册
		xerror.Assert(app.Command(cmd.Name) != nil, "command(%s) already exists", cmd.Name)
		app.Commands = append(app.Commands, cmd)
	}

	sort.Sort(cli.FlagsByName(app.Flags))
	sort.Sort(cli.CommandsByName(app.Commands))
	xerror.Panic(app.Run(os.Args))
}
