package debug

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/mux"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/plugins/syncx"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
)

func init() {
	var logs = logz.Component(Name)

	var openWeb bool
	plugin.Register(&plugin.Base{
		Name: Name,
		OnFlags: func() types.Flags {
			return types.Flags{
				&cli.BoolFlag{
					Name:        "debug.web",
					Value:       openWeb,
					Destination: &openWeb,
					Usage:       "open web browser with debug",
				},
				&cli.StringFlag{
					Name:        "debug.addr",
					Destination: &runenv.DebugAddr,
					Usage:       "debug server http address",
					Value:       runenv.DebugAddr,
					EnvVars:     types.EnvOf("lava-debug-addr"),
				},
			}
		},
		OnInit: func(p plugin.Process) {
			InitView()
			var server = &http.Server{Addr: runenv.DebugAddr, Handler: mux.Mux()}
			p.AfterStart(func() {
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

			p.AfterStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					logs.WithErr(err).Error("Server [debug] Shutdown Error")
				}
			})
		},
	})
}
