package debug

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/pubgo/dix"
	"github.com/pubgo/lug/abc"
	cb "github.com/pubgo/lug/builder/chi"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/vars"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/pflag"
)

func On(fn func(mux *abc.DebugMux)) { xerror.Panic(dix.Dix(fn)) }
func init()                         { plugin.Register(plg) }

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

		xerror.Panic(dix.Dix((*abc.DebugMux)(srv.Mux)))

		var server = &http.Server{Addr: Addr, Handler: srv}
		entry.Parse(ent, func(ent entry.Entry) {
			ent.BeforeStart(func() {
				xerror.Exit(fx.GoDelay(time.Second, func() {
					xlog.Infof("Server [debug] Listening on http://localhost%s", Addr)
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						xlog.Error("Server [debug] Listen Error", xlog.Any("err", err))
						return
					}

					xlog.Info("Server [debug] Closed OK")
				}))

			})

			ent.BeforeStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					xlog.Error("Server [debug] Shutdown Error", xlog.Any("err", err))
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
