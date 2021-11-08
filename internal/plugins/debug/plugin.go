package debug

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/mux"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runenv"
)

func init() {
	var logs = logz.New(Name)

	var openWeb bool
	plugin.Register(&plugin.Base{
		Name: Name,
		OnFlags: func() []cli.Flag {
			return []cli.Flag{
				&cli.BoolFlag{
					Name:        "debug.web",
					Value:       openWeb,
					Destination: &openWeb,
					Usage:       "open web browser with debug mode",
				},
			}
		},
		OnInit: func() {
			InitView()
			var server = &http.Server{Addr: runenv.DebugAddr, Handler: mux.Mux()}
			entry.AfterStart(func() {
				xerror.Assert(netutil.CheckPort("tcp4", runenv.DebugAddr), "server: %s already exists", runenv.DebugAddr)
				syncx.GoDelay(func() {
					logs.Infof("Server [debug] Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runenv.DebugAddr))
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						logs.WithErr(err).Error("Server [debug] Listen Error")
						return
					}
					logs.Info("Server [debug] Closed OK")
				})

				if openWeb {
					syncx.GoDelay(func() {
						xerror.Panic(browser.OpenURL(fmt.Sprintf("http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runenv.DebugAddr))))
					})
				}
			})

			entry.AfterStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					logs.WithErr(err).Error("Server [debug] Shutdown Error")
				}
			})
		},
	})
}
