package main

import (
	"os"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmd/cmds/hello"
	"github.com/pubgo/lava/cmd/cmds/logs"
	"github.com/pubgo/lava/cmd/cmds/mage"
	"github.com/pubgo/lava/cmd/cmds/protoc"
	"github.com/pubgo/lava/cmd/cmds/swagger"
	"github.com/pubgo/lava/cmd/cmds/trace"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/version"
)

func main() {
	runenv.Project = "lava"

	xerror.Exit((&cli.App{
		Name:    runenv.Project,
		Version: version.Version,
		Commands: cli.Commands{
			trace.Cmd(),
			protoc.Cmd(),
			swagger.Cmd,
			logs.Cmd,
			hello.Cmd,
			mage.Cmd,
		},
	}).Run(os.Args))
}
