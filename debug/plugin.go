package debug

import (
	"context"
	"fmt"
	"net/http"

	"github.com/pubgo/dix"
	cb "github.com/pubgo/lug/builder/chi"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func On(fn func(mux *types.DebugMux)) { xerror.Panic(dix.Dix(fn)) }
func init()                           { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnFlags: func(flags *pflag.FlagSet) {
		flags.StringVar(&Addr, "da", Addr, "debug server addr")
	},
	OnInit: func(ent interface{}) {
		var srv = &Mux{Cfg: cb.GetDefaultCfg(), srv: cb.New()}
		_ = config.Decode(Name, &srv)

		var builder = cb.New()
		xerror.Panic(builder.Build(srv.Cfg))
		srv.Mux = builder.Get()

		xerror.Panic(dix.Dix((*types.DebugMux)(srv.Mux)))

		var server = &http.Server{Addr: Addr, Handler: srv}
		entry.Parse(ent, func(ent entry.Entry) {
			ent.BeforeStart(func() {
				fx.GoDelay(func() {
					logs.Infof("Server [debug] Listening on http://localhost%s", Addr)
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						logs.Error("Server [debug] Listen Error", zap.Any("err", err))
						return
					}

					logs.Info("Server [debug] Closed OK")
				})

			})

			ent.BeforeStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					logs.Error("Server [debug] Shutdown Error", zap.Any("err", err))
				}
			})
		})

		vars.Watch(Name+"_rest_router", func() interface{} {
			var dt []string
			for _, r := range srv.Routes() {
				dt = append(dt, fmt.Sprintf("http://localhost%s%s", Addr, r.Pattern))
			}

			return dt
		})
	},
}
