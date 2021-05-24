package vars

import (
	"expvar"
	"fmt"
	"strconv"
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

		var t = 1
		if len(args) > 0 {
			var a1, err = strconv.Atoi(args[0])
			if err != nil {
				t = a1
			}
		}

		for {
			expvar.Do(func(val expvar.KeyValue) {
				fmt.Println(val.Key, val.Value.String())
			})
			fmt.Print("\n\n\n")

			time.Sleep(time.Duration(t) * time.Second)
		}
	},
}
