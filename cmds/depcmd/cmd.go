package depcmd

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	cli "github.com/urfave/cli/v3"

	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/pkg/cmdx"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "dep",
		Usage: "Print the dependency package information",
		Description: cmdx.ExampleFmt(
			"lava dep",
			"lava dep json",
			"lava dep t"),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()

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
				dt := assert.Must1(json.MarshalIndent(runmode.GetVersion(), "", "\t"))
				fmt.Println(string(dt))
			case "json":
				dt := assert.Must1(json.MarshalIndent(info, "", "\t"))
				fmt.Println(string(dt))
			case "table", "tb", "t":
				table := tablewriter.NewWriter(os.Stdout)
				table.SetHeader([]string{"path", "Version", "Replace"})
				table.Append([]string{info.Main.Path, version.Version(), replace(info.Main.Replace)})

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
