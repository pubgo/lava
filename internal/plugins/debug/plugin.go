package debug

import (
	"context"
	"fmt"
	"github.com/pubgo/lava/pkg/syncx"
	"net/http"

	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	"github.com/urfave/cli/v2"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/mux"
	"github.com/pubgo/lava/pkg/env"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
)

func init() {
	var logs = logging.Component(Name)

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
					Destination: &runtime.DebugAddr,
					Usage:       "debug server http address",
					Value:       runtime.DebugAddr,
					EnvVars:     env.KeyOf("lava-debug-addr"),
				},
			}
		},
		OnInit: func(p plugin.Process) {
			InitView()
			var server = &http.Server{Addr: runtime.DebugAddr, Handler: mux.Mux()}
			p.AfterStart(func() {
				xerror.Assert(netutil.CheckPort("tcp4", runtime.DebugAddr), "server: %s already exists", runtime.DebugAddr)
				syncx.GoDelay(func() {
					logs.S().Infof("Server [debug] Listening on http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.DebugAddr))
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						logs.WithErr(err).Error("Server [debug] Listen Error")
						return
					}
					logs.L().Info("Server [debug] Closed OK")
				})

				if openWeb {
					syncx.GoDelay(func() {
						xerror.Panic(browser.OpenURL(fmt.Sprintf("http://%s:%d", netutil.GetLocalIP(), netutil.MustGetPort(runtime.DebugAddr))))
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
