package depcmd

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/pretty"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
	"github.com/pubgo/funk/v2/version"
	cli "github.com/urfave/cli/v3"

	"github.com/pubgo/lava/v2/pkg/cmdutil"
)

func New(di *dix.Dix) *cli.Command {
	return &cli.Command{
		Name:  "dep",
		Usage: "Print the dependency package information",
		Description: cmdutil.ExampleFmt(
			"lava dep",
			"lava dep json",
			"lava dep t"),
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			info, ok := debug.ReadBuildInfo()
			if !ok {
				return nil
			}

			var typ string
			if command.NArg() > 0 {
				typ = command.Args().First()
			}

			switch typ {
			case "":
				pretty.Println(info)
			case "sys":
				pretty.Println(running.GetSysInfo())
			case "table", "tb", "t":
				table := tablewriter.NewWriter(os.Stdout)
				table.Header([]string{"path", "Version", "Replace"})
				assert.Must(table.Append([]string{info.Main.Path, version.Version(), replace(info.Main.Replace)}))

				for _, dep := range info.Deps {
					assert.Must(table.Append([]string{dep.Path, dep.Version, replace(dep.Replace)}))
				}
				assert.Must(table.Render())
			case "di":
				fmt.Println(di.Graph().Objects)
				fmt.Println(di.Graph().Providers)
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
