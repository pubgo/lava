package gcnotifier

import (
	"context"
	"github.com/pubgo/lava/core/runmode"

	"github.com/CAFxX/gcnotifier"
	"go.uber.org/fx"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/pkg/syncx"
)

var Name = "gc"
var logs = logging.Component(Name)

func Module() fx.Option {
	return fx.Invoke(func(r lifecycle.Lifecycle) {
		if runmode.IsProd() || runmode.IsRelease() {
			return
		}

		r.AfterStops(func() {
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
	})
}
