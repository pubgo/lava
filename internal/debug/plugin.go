package debug

import (
	"context"
	"expvar"
	"fmt"
	"net/http"

	cb "github.com/pubgo/lug/builder/chi"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/pkg/gutil"
	"github.com/pubgo/lug/pkg/netutil"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/types"
	"github.com/pubgo/lug/vars"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"github.com/spf13/pflag"
	"go.uber.org/zap"
)

func On(fn func(mux *types.DebugMux)) { xerror.Exit(dix.Provider(fn)) }
func init()                           { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnFlags: func(flags *pflag.FlagSet) {
		flags.StringVar(&Addr, "da", Addr, "debug server addr")
	},

	OnInit: func(ent entry.Entry) {
		var srv = &Mux{Cfg: cb.GetDefaultCfg(), srv: cb.New()}
		_ = config.Decode(Name, &srv)

		var builder = cb.New()
		xerror.Panic(builder.Build(srv.Cfg))
		srv.Mux = builder.Get()

		xerror.Exit(dix.Provider(srv.Mux))

		var server = &http.Server{Addr: Addr, Handler: srv}
		ent.BeforeStart(func() {
			expvar.Do(func(kv expvar.KeyValue) {
				var val = kv.Value
				srv.Mux.Get(fmt.Sprintf("/debug/vars/%s", kv.Key), func(writer http.ResponseWriter, request *http.Request) {
					writer.Header().Set("Content-Type", "application/json")
					xerror.PanicErr(writer.Write(strutil.ToBytes(val.String())))
				})
			})

			xerror.Assert(netutil.ScanPort("tcp4", Addr), "server: %s already exists", Addr)

			fx.GoDelay(func() {
				logs.Sugar().Infof("Server [debug] Listening on http://localhost:%s", gutil.GetPort(Addr))
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

		vars.Watch(Name+"_rest_router", func() interface{} {
			var dt []string
			for _, r := range srv.Routes() {
				dt = append(dt, fmt.Sprintf("http://localhost%s%s", Addr, r.Pattern))
			}

			return dt
		})
	},
}
