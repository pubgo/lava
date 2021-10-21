package debug

import (
	"context"
	"expvar"
	"fmt"
	"github.com/pubgo/lava/mux"
	"net/http"

	varView "github.com/go-echarts/statsview/expvar"
	"github.com/pkg/browser"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"

	"github.com/pubgo/lava/logger"
	"github.com/pubgo/lava/logz"
	"github.com/pubgo/lava/pkg/lavax"
	"github.com/pubgo/lava/pkg/netutil"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runenv"
)

func init() {
	var logs = logz.Named(Name)

	var openWeb bool
	plugin.Register(&plugin.Base{
		Name: Name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.BoolVar(&openWeb, "web", openWeb, "open web browser")
		},
		OnInit: func(ent plugin.Entry) {
			InitView()

			expvar.Do(func(value expvar.KeyValue) {
				AddView(varView.NewExpvarViewer(value.Key))
			})

			serveMux := GetDefaultServeMux()
			for k, v := range serveMux.M {
				mux.Get(k, v.H.ServeHTTP)
			}

			var server = &http.Server{Addr: runenv.DebugAddr, Handler: mux.Mux()}
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

			ent.AfterStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					logs.Error("Server [debug] Shutdown Error", logger.WithErr(err))
				}
			})
		},
	})
}
