package healthy

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"

	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/runenv"
)

var Cmd = &cobra.Command{
	Use:   "health",
	Short: "health check",
	Example: gutil.CmdExample(
		"lug health",
		"lug health localhost:8081",
	),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		var addr = runenv.DebugAddr
		if len(args) > 0 {
			addr = args[0]
		}

		var resp, err = http.Get(fmt.Sprintf("http://localhost:%s/health", gutil.GetPort(addr)))
		xerror.Panic(err)
		xerror.Assert(resp.StatusCode != http.StatusOK, "health check")
		_, _ = io.Copy(os.Stdout, resp.Body)
	},
}
