package healthcmd

import (
	"fmt"
	netutil2 "github.com/pubgo/lava/internal/pkg/netutil"
	"io"
	"net/http"
	"os"

	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"
)

func Cmd() *cli.Command {
	return &cli.Command{
		Name:  "health",
		Usage: "health check",
		Description: flagx.ExampleFmt(
			"lava health",
			"lava health localhost:8080",
		),
		Action: func(ctx *cli.Context) error {
			defer xerror.RecoverAndExit()

			var addr = ":8080"
			if ctx.NArg() > 0 {
				addr = ctx.Args().First()
			}

			var resp, err = http.Get(fmt.Sprintf("http://%s:%d/health", netutil2.GetLocalIP(), netutil2.MustGetPort(addr)))
			xerror.Panic(err)
			xerror.Assert(resp.StatusCode != http.StatusOK, "health check")
			_, _ = io.Copy(os.Stdout, resp.Body)
			return nil
		},
	}
}
