package healthy

import (
	"fmt"
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
	Run: func(cmd *cobra.Command, args []string) {
		defer xerror.RespExit()

		var addrs = strings.Split(debug.Addr, ":")
		var resp, err = http.Get(fmt.Sprintf("http://localhost:%s/health", addrs[len(addrs)-1]))
		xerror.Panic(err)
		xerror.Assert(resp.StatusCode != http.StatusOK, "health check")
		_, _ = io.Copy(os.Stdout, resp.Body)
	},
}
