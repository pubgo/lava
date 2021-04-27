package vars

import (
	"expvar"
	"fmt"
	"strings"
	"time"

	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

var trim = strings.TrimSpace
var Cmd = &cobra.Command{
	Use:   "var",
	Short: "Print the expvar information",
	Example: trim(`
lug var
lug var 1s`),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		var t = time.Second
		if len(args) > 0 {
			t, _ = time.ParseDuration(args[0])
		}

		for {
			expvar.Do(func(val expvar.KeyValue) {
				fmt.Println(val.Key, val.Value.String())
			})

			time.Sleep(t)
		}
	},
}
