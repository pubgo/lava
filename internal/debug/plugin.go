package debug

import (
	"context"
	"net/http"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/logger"
	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/pkg/netutil"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/runenv"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"go.uber.org/zap"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnInit: func(ent entry.Entry) {
		var server = &http.Server{Addr: runenv.DebugAddr}
		ent.AfterStart(func() {
			xerror.Assert(netutil.ScanPort("tcp4", runenv.DebugAddr), "server: %s already exists", runenv.DebugAddr)
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
}
