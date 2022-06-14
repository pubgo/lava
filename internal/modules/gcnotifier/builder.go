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
var logs = logging.Component(Name)

func init() {
	dix.Register(func() lifecycle.Handler {
		if runmode.IsProd() || runmode.IsRelease() {
			return nil
		}

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
