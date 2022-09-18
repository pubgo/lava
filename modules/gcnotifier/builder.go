package gcnotifier

import (
	"context"

	"github.com/CAFxX/gcnotifier"
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/result"
	"github.com/pubgo/funk/syncx"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/logging"
)

var Name = "gc"

func init() {
	di.Provide(func(log *logging.Logger) lifecycle.Handler {
		if !runmode.IsDebug {
			return nil
		}

		var logs = logging.ModuleLog(log, Name)
		return func(lc lifecycle.Lifecycle) {
			lc.AfterStart(func() {
				var cancel = syncx.GoCtx(func(ctx context.Context) result.Error {
					var gc = gcnotifier.New()
					defer gc.Close()

					// TODO handler

					for {
						select {
						case <-gc.AfterGC():
							logs.L().Info("gc notify")
						case <-ctx.Done():
							return result.Error{}
						}
					}
				})
				lc.BeforeStop(cancel)
			})
		}
	})
}
