package servefilecmd

import (
	"context"
	"fmt"
	"os"

	"github.com/pubgo/funk/assert"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/recovery"
	"github.com/pubgo/funk/running"
	"github.com/pubgo/funk/v2/result"
	cli "github.com/urfave/cli/v3"
	"github.com/valyala/fasthttp"
)

func New() *cli.Command {
	return &cli.Command{
		Name:  "servefile",
		Usage: "serve `pwd` via http at *:8080",
		Flags: []cli.Flag{},
		Action: func(ctx context.Context, command *cli.Command) error {
			defer recovery.Exit()

			wd := result.Wrap(os.Getwd()).Must()

			port := running.HttpPort
			log.Info().Msgf("file dir: %s", wd)
			log.Info().Msgf("http://localhost:%v", port)

			fs := &fasthttp.FS{
				Root:               wd,
				IndexNames:         []string{"index.html"},
				GenerateIndexPages: true,
				Compress:           false,
				AcceptByteRange:    true,
				//PathRewrite:     fasthttp.NewVHostPathRewriter(0),
			}

			s := &fasthttp.Server{
				Handler: fs.NewRequestHandler(),
				Logger:  log.NewStd(log.GetLogger("servefile")),
			}
			go func() {
				assert.Must(s.ListenAndServe(fmt.Sprintf(":%v", port)))
			}()

			<-ctx.Done()
			return s.ShutdownWithContext(ctx)
		},
	}
}
