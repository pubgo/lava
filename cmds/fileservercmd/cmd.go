package fileservercmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pubgo/funk/v2/assert"
	"github.com/pubgo/funk/v2/log"
	"github.com/pubgo/funk/v2/recovery"
	"github.com/pubgo/funk/v2/result"
	"github.com/pubgo/redant"
	"github.com/valyala/fasthttp"

	"github.com/pubgo/lava/v2/core/running"
	"github.com/pubgo/lava/v2/pkg/netutil"
)

func New() *redant.Command {
	return &redant.Command{
		Use:   "fileserver <dir>",
		Short: "serve `pwd` via http at *:8080 or <http-port>",
		Handler: func(ctx context.Context, command *redant.Invocation) error {
			defer recovery.Exit()

			wd := result.Wrap(os.Getwd()).Unwrap()
			if len(command.Args) > 0 {
				wd = command.Args[0]
			}

			port := running.HttpPort.Value()
			log.Info().Msgf("file dir: %s", wd)
			log.Info().Msgf("http://localhost:%v", port)

			fs := &fasthttp.FS{
				Root:               wd,
				IndexNames:         []string{"index.html"},
				GenerateIndexPages: true,
				Compress:           false,
				AcceptByteRange:    true,
				// PathRewrite:     fasthttp.NewVHostPathRewriter(0),
			}

			s := &fasthttp.Server{
				Handler: fs.NewRequestHandler(),
				Logger:  log.NewStd(log.GetLogger("fileserver")),
			}
			go func() {
				assert.Exit(netutil.SkipServerClosedError(s.ListenAndServe(fmt.Sprintf(":%v", port))))
			}()

			<-ctx.Done()
			return s.ShutdownWithContext(ctx)
		},
	}
}
