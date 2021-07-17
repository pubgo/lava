package healthy

import (
	"fmt"
	"github.com/pubgo/lug/pkg/gutil"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/pubgo/lug/internal/debug"
	"github.com/pubgo/xerror"
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "health",
	Short: "health check",
	Example: gutil.CmdExample(
		`lug health`,
		"lug health localhost:8081",
	),
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		var addr = debug.Addr
		if len(args) > 0 {
			addr = args[0]
		}

		var addrs = strings.Split(addr, ":")
		var resp, err = http.Get(fmt.Sprintf("http://localhost:%s/health", addrs[len(addrs)-1]))
		xerror.Panic(err)
		xerror.Assert(resp.StatusCode != http.StatusOK, "health check")
		_, _ = io.Copy(os.Stdout, resp.Body)
	},
}
