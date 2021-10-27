package healthy

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"

	"github.com/pubgo/lava/pkg/clix"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/runenv"
)

var Cmd = &cobra.Command{
	Use:   "health",
	Short: "health check",
	Example: clix.ExampleFmt(
		"lava health",
		"lava health localhost:8081",
	),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		var addr = runenv.DebugAddr
		if len(args) > 0 {
			addr = args[0]
		}

		var resp, err = http.Get(fmt.Sprintf("http://localhost:%s/health", lavax.GetPort(addr)))
		xerror.Panic(err)
		xerror.Assert(resp.StatusCode != http.StatusOK, "health check")
		_, _ = io.Copy(os.Stdout, resp.Body)
	},
}
