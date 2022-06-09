package vercmd

import (
	"encoding/json"
	"fmt"
	"github.com/pubgo/lava/core/runmode"
	"os"
	"runtime/debug"

	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/clix"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/version"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:    "version",
		Aliases: typex.StrOf("v"),
		Usage:   "Print the dependency package information",
		Description: clix.ExampleFmt(
			"lava version",
			"lava version json",
			"lava version t"),
		Action: func(ctx *cli.Context) error {
			defer xerror.RecoverAndExit()

			info, ok := debug.ReadBuildInfo()
			if !ok {
				return nil
			}

			var typ string

			if ctx.NArg() > 0 {
				typ = ctx.Args().First()
			}

			switch typ {
			case "":
				dt, err := json.MarshalIndent(runmode.GetVersion(), "", "\t")
				xerror.Panic(err)
				fmt.Println(string(dt))
			case "json":
				dt, err := json.MarshalIndent(info, "", "\t")
				xerror.Panic(err)
				fmt.Println(string(dt))
			case "table", "tb", "t":
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"path", "Version", "Replace"})
				table.Append([]string{info.Main.Path, version.Version, replace(info.Main.Replace)})

				for _, dep := range info.Deps {
					table.Append([]string{dep.Path, dep.Version, replace(dep.Replace)})
				}
				table.Render()
			}
			return nil
		},
	}
}

func replace(dep *debug.Module) string {
	if dep == nil {
		return ""
	}

	return fmt.Sprintf("%s:%s", dep.Path, dep.Version)
}
