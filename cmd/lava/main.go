package main

import (
	"fmt"
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/version"
)

func main() {
	var app = &cli.App{
		Name:    runmode.Project,
		Version: version.Version(),
		Action: func(context *cli.Context) error {
			fmt.Println(version.Domain())
			fmt.Println(version.Version())
			fmt.Println(version.CommitID())
			fmt.Println(version.Project())
			return nil
		},
		Commands: cli.Commands{},
	}
	assert.Exit(app.Run(os.Args))
}
