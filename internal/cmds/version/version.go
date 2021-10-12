package version

import (
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pubgo/lava/version"
	"github.com/pubgo/x/typex"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

var trim = strings.TrimSpace
var Cmd = &cobra.Command{
	Use:     "version",
	Aliases: typex.StrOf("v"),
	Short:   "Print the dependency package information",
	Example: trim(`
lava version
lava version json
lava version t`),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		info, ok := debug.ReadBuildInfo()
		if !ok {
			return
		}

		var typ string

		if len(args) > 0 {
			typ = args[0]
		}

		switch typ {
		case "":
			dt, err := json.MarshalIndent(version.GetVer(), "", "\t")
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
	},
}

func replace(dep *debug.Module) string {
	if dep == nil {
		return ""
	}

	return fmt.Sprintf("%s:%s", dep.Path, dep.Version)
}
