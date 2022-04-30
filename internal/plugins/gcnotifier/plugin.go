package gcnotifier

import (
	"context"
	"github.com/pubgo/lava/inject"
	"go.uber.org/fx"

	"github.com/CAFxX/gcnotifier"

	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/syncx"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
)

var Name = "gc"
var logs = logging.Component(Name)

func init() {
	if runtime.IsProd() || runtime.IsRelease() {
		return
	}

	inject.Register(fx.Invoke(func(srv service.Service) {
		srv.AfterStops(func() {
			syncx.GoCtx(func(ctx context.Context) {
				var gc = gcnotifier.New()
				defer gc.Close()

				// TODO handler

				for {
					select {
					case <-gc.AfterGC():
						logs.L().Info("gc notify")
					case <-ctx.Done():
						return
					}
				}
			})
		})
	}))
}
