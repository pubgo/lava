package main

import (
	"os"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/cmd/lava/cmds/protoc"
	"github.com/pubgo/lava/version"
)

func main() {
	var app = &cli.App{
		Name:     "prototool",
		Version:  version.Version,
		Commands: cli.Commands{protoc.Cmd()},
	}
	xerror.Exit(app.Run(os.Args))
}