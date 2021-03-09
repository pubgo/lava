package mux

import (
	"context"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/internal/golug_run"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"github.com/spf13/pflag"
)

func init() {
	server := &http.Server{Addr: addr, Handler: app}
	plugin.Register(&plugin.Base{
		Name: Name,
		OnFlags: func(flags *pflag.FlagSet) {
			flags.StringVar(&server.Addr, "da", addr, "debug addr")
		},

		OnInit: func(ent interface{}) {
			cfg := fiber.New().Config()

			config.Decode(Name, &cfg)

			golug_run.AfterStart(func() {
				xerror.Exit(fx.GoDelay(time.Second, func() {
					if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
						xlog.Error("Server Listen Error", xlog.Any("err", err))
					}

					xlog.Info("Server [http] Closed OK")
				}))

				xlog.Infof("Server [http] Listening on http://%s", addr)
			})

			golug_run.BeforeStop(func() {
				if err := server.Shutdown(context.Background()); err != nil {
					xlog.Error("Server Shutdown Error", xlog.Any("err", err))
				}
			})
		},
	})
}
