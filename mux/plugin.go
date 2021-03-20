package mux

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/dix"
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func On(fn func(app *chi.Mux)) { xerror.Panic(dix.Dix(fn)) }

func onInit(ent interface{}) {
	if !config.Decode(Name, &cfg) {
		return
	}

	var app = cfg.Build()
	var addr = fmt.Sprintf(":%d", config.DebugPort)
	var server = &http.Server{Addr: addr, Handler: app}
	xerror.Panic(dix.Dix(app))

	golug_run.AfterStart(func() {
		xerror.Exit(fx.GoDelay(time.Second, func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				xlog.Error("Server [mux] Listen Error", xlog.Any("err", err))
			}

			xlog.Info("Server [mux] Closed OK")
		}))
		xlog.Infof("Server [mux] Listening on http://localhost%s", addr)
	})

	golug_run.BeforeStop(func() {
		if err := server.Shutdown(context.Background()); err != nil {
			xlog.Error("Server Shutdown Error", xlog.Any("err", err))
		}
	})

	expvar.Publish(Name+"_rest_router", expvar.Func(func() interface{} {
		if app == nil {
			return nil
		}

		return fmt.Sprintf("%#v\n", app.Routes())
	}))
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
