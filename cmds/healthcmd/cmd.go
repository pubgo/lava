package healthcmd

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/buildinfo/version"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/redant"

	"github.com/pubgo/lava/v2/pkg/cliutil"
	"github.com/pubgo/lava/v2/pkg/netutil"
)

func New() *redant.Command {
	return &redant.Command{
		Use:   "health",
		Short: cliutil.UsageDesc("%s health check", version.Project()),
		Handler: func(ctx context.Context, command *redant.Invocation) error {
			defer recovery.Exit()

			addr := ":8080"
			if len(command.Args) > 0 {
				addr = command.Args[0]
			}

			resp := assert.Must1(http.Get(fmt.Sprintf("http://%s:%d/health", netutil.GetLocalIP(), netutil.MustGetPort(addr))))
			assert.If(resp.StatusCode != http.StatusOK, "health check")
			_, _ = io.Copy(os.Stdout, resp.Body)
			return nil
		},
	}
}
