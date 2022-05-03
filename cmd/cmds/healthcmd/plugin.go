package healthcmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/pkg/clix"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/runtime"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "health",
		Usage: "health check",
		Description: clix.ExampleFmt(
			"lava health",
			"lava health localhost:8080",
		),
		Action: func(ctx *cli.Context) error {
			defer xerror.RespExit()

			var addr = runtime.Addr
			if ctx.NArg() > 0 {
				addr = ctx.Args().First()
			}

			var resp, err = http.Get(fmt.Sprintf("http://%s:%d/health", netutil.GetLocalIP(), netutil.MustGetPort(addr)))
			xerror.Panic(err)
			xerror.Assert(resp.StatusCode != http.StatusOK, "health check")
			_, _ = io.Copy(os.Stdout, resp.Body)
			return nil
		},
	}
}
