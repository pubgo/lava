package debug

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pkg/browser"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/internal/logz"
	"github.com/pubgo/lava/mux"
	"github.com/pubgo/lava/pkg/lavax"
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
		OnFlags: func(flags *pflag.FlagSet) {
			flags.BoolVar(&openWeb, "web", openWeb, "open web browser")
		},
		OnInit: func(ent plugin.Entry) {
			InitView()

			serveMux := GetDefaultServeMux()
			for k, v := range serveMux.M {
				mux.HandleFunc(k, v.H.ServeHTTP)
			}

			var server = &http.Server{Addr: runenv.DebugAddr, Handler: mux.Mux()}
			ent.AfterStart(func() {
				xerror.Assert(netutil.CheckPort("tcp4", runenv.DebugAddr), "server: %s already exists", runenv.DebugAddr)
				syncx.GoDelay(func() {
					logs.Infof("Server [debug] Listening on http://localhost:%s", lavax.GetPort(runenv.DebugAddr))
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						logs.WithErr(err).Error("Server [debug] Listen Error")
						return
					}
					logs.Info("Server [debug] Closed OK")
				})

				if openWeb {
					syncx.GoDelay(func() {
						xerror.Panic(browser.OpenURL(fmt.Sprintf("http://localhost:%s", lavax.GetPort(runenv.DebugAddr))))
					})
				}
			})

			ent.AfterStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					logs.WithErr(err).Error("Server [debug] Shutdown Error")
				}
			})
		},
	})
}
