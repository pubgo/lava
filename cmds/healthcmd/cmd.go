package healthcmd

import (
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/version"
	"github.com/urfave/cli/v3"

	"github.com/pubgo/lava/pkg/cmdutil"
	"github.com/pubgo/lava/pkg/netutil"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "health",
		Usage: cmdutil.UsageDesc("%s health check", version.Project()),
		Description: cmdutil.ExampleFmt(
			"lava health",
			"lava health localhost:8080",
		),
		Action: func(ctx *cli.Context) error {
			defer recovery.Exit()

			var addr = ":8080"
			if ctx.NArg() > 0 {
				addr = ctx.Args().First()
			}

			var resp = assert.Must1(http.Get(fmt.Sprintf("http://%s:%d/health", netutil.GetLocalIP(), netutil.MustGetPort(addr))))
			assert.If(resp.StatusCode != http.StatusOK, "health check")
			_, _ = io.Copy(os.Stdout, resp.Body)
			return nil
		},
	}
}
