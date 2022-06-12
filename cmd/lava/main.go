package main

import (
	"os"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/version"
)

func main() {
	var app = &cli.App{
		Name:     runmode.Project,
		Version:  version.Version,
		Commands: cli.Commands{},
	}
	xerror.Exit(app.Run(os.Args))
}
