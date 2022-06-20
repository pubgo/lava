package gcnotifier

import (
	"context"

	"github.com/CAFxX/gcnotifier"
	"github.com/pubgo/dix"

	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/core/runmode"
	"github.com/pubgo/lava/internal/pkg/syncx"
	"github.com/pubgo/lava/logging"
)

var Name = "gc"

func init() {
	dix.Provider(func(log *logging.Logger) lifecycle.Handler {
		if runmode.IsProd() || runmode.IsRelease() {
			return nil
		}

		var logs = logging.ModuleLog(log, Name)
		return func(lc lifecycle.Lifecycle) {
			lc.AfterStarts(func() {
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
		}
	})
}
