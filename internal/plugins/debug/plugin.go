package debug

import (
	"context"
	"expvar"
	"fmt"
	varView "github.com/go-echarts/statsview/expvar"
	"net/http"

	"github.com/pkg/browser"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/debug"
	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runenv"
)

func init() {
	var logs = logz.Named(debug.Name)

	var openWeb bool
	plugin.Register(&plugin.Base{
		Name: debug.Name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.BoolVar(&openWeb, "web", openWeb, "open web browser")
		},
		OnInit: func(ent plugin.Entry) {
			debug.InitView()

			expvar.Do(func(value expvar.KeyValue) {
				debug.AddView(varView.NewExpvarViewer(value.Key))
			})

			serveMux := debug.GetDefaultServeMux()
			for k, v := range serveMux.M {
				debug.Get(k, v.H.ServeHTTP)
			}

			var server = &http.Server{Addr: runenv.DebugAddr, Handler: debug.Mux()}
			ent.AfterStart(func() {
				xerror.Assert(netutil.CheckPort("tcp4", runenv.DebugAddr), "server: %s already exists", runenv.DebugAddr)
				fx.GoDelay(func() {
					logs.Infof("Server [debug] Listening on http://localhost:%s", lavax.GetPort(runenv.DebugAddr))
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						logs.Error("Server [debug] Listen Error", logger.WithErr(err))
						return
					}
					logs.Info("Server [debug] Closed OK")
				})
				if openWeb {
					fx.Go(func(ctx context.Context) {
						xerror.Panic(browser.OpenURL(fmt.Sprintf("http://localhost:%s", lavax.GetPort(runenv.DebugAddr))))
					})
				}
			})

			ent.BeforeStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					logs.Error("Server [debug] Shutdown Error", logger.WithErr(err))
				}
			})
		},
	})
}
