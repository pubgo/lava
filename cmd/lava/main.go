package main

import (
	"os"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmd/lava/cmds/initCmd"
	"github.com/pubgo/lava/cmd/lava/cmds/mage"
	"github.com/pubgo/lava/cmd/lava/cmds/protoc"
	"github.com/pubgo/lava/cmd/lava/cmds/swagger"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/version"
)

func main() {
	var app = &cli.App{
		Name:    runtime.Project,
		Version: version.Version,
		Commands: cli.Commands{
			initCmd.Cmd(),
			protoc.Cmd(),
			swagger.Cmd,
			mage.Cmd,
		},
	}
	xerror.Exit(app.Run(os.Args))
}
