package mux

import (
	"context"
	"net/http"
	"time"

	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/pflag"
)

func onInit(ent interface{}) {
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
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVar(&server.Addr, "da", addr, "debug addr")
		},
	})
}
