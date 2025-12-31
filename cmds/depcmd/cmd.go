package depcmd

import (
	"context"
	"fmt"
	"os"
	"runtime/debug"

	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/dix/v2"
	"github.com/pubgo/redant"
	"github.com/samber/lo"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/pretty"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/running"
)

func New(di *dix.Dix) *redant.Command {
	return &redant.Command{
		Use:   "dep",
		Short: "Print the dependency package information",
		Handler: func(ctx context.Context, i *redant.Invocation) error {
			defer recovery.Exit()

			info, ok := debug.ReadBuildInfo()
			if !ok {
				return nil
			}

			typ := lo.FirstOrEmpty(i.Args)
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
