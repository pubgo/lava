package mux

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/pubgo/dix"
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
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

	ent1 := ent.(entry.Entry)
	ent1.BeforeStart(func() {
		xerror.Exit(fx.GoDelay(time.Second, func() {
			if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				xlog.Error("Server [mux] Listen Error", xlog.Any("err", err))
			}

			xlog.Info("Server [mux] Closed OK")
		}))
		xlog.Infof("Server [mux] Listening on http://localhost%s", addr)
	})

	ent1.BeforeStop(func() {
		if err := server.Shutdown(context.Background()); err != nil {
			xlog.Error("Server Shutdown Error", xlog.Any("err", err))
		}
	})

	type Route struct {
		Pattern   string
		Handlers  map[string]bool
		SubRoutes []Route
	}

	var getRoutes func(routes []chi.Route) []Route
	getRoutes = func(routes []chi.Route) []Route {
		if len(routes) == 0 {
			return nil
		}

		var routeList []Route
		for _, r := range app.Routes() {
			rr := Route{Pattern: r.Pattern, Handlers: make(map[string]bool)}

			for k := range r.Handlers {
				rr.Handlers[k] = true
			}

			//if r.SubRoutes != nil {
			//	rr.SubRoutes = getRoutes(r.SubRoutes.Routes())
			//}

			routeList = append(routeList, rr)
		}
		return routeList
	}

	expvar.Publish(Name+"_rest_router", expvar.Func(func() interface{} {
		if app == nil {
			return nil
		}

		fmt.Printf("%#v\n", getRoutes(app.Routes()))
		return getRoutes(app.Routes())
	}))
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
