package debug

import (
	"context"
	"net/http"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"

	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/mux"
	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/pkg/netutil"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			serveMux := mux.GetDefaultServeMux()
			for k, v := range serveMux.M {
				mux.Get(k, v.H.ServeHTTP)
			}

			var server = &http.Server{Addr: runenv.DebugAddr, Handler: mux.Mux()}
			ent.AfterStart(func() {
				xerror.Assert(netutil.CheckPort("tcp4", runenv.DebugAddr), "server: %s already exists", runenv.DebugAddr)
				fx.GoDelay(func() {
					zap.S().Infof("Server [debug] Listening on http://localhost:%s", gutil.GetPort(runenv.DebugAddr))
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						zap.L().Error("Server [debug] Listen Error", logger.Err(err))
						return
					}
					zap.L().Info("Server [debug] Closed OK")
				})
			})

			ent.BeforeStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					zap.L().Error("Server [debug] Shutdown Error", logger.Err(err))
				}
			})
		},
	})
}
