package vars

import (
	"expvar"
	"fmt"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
	"strings"
)

var trim = strings.TrimSpace
var Cmd = &cobra.Command{
	Use:     "var",
	Short:   "Print the expvar information",
	Example: trim(`lug var`),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		expvar.Do(func(val expvar.KeyValue) {
			fmt.Println(val.Key, val.Value.String())
		})
		fmt.Print("\n\n\n")
	},
}
