package gcnotifier

import (
	"context"

	"github.com/CAFxX/gcnotifier"
	"github.com/pubgo/dix/di"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/lifecycle"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/lava/core/runmode"
)

var Name = "gc"

func init() {
	di.Provide(func(log log.Logger) lifecycle.Handler {
		if !runmode.IsDebug {
			return nil
		}

		var logs = log.WithName(Name)
		return func(lc lifecycle.Lifecycle) {
			lc.AfterStart(func() {
				var cancel = async.GoCtx(func(ctx context.Context) error {
					var gc = gcnotifier.New()
					defer gc.Close()

					// TODO handler

					for {
						select {
						case <-gc.AfterGC():
							logs.Info().Msg("gc notify")
						case <-ctx.Done():
							return nil
						}
					}
				})
				lc.BeforeStop(cancel)
			})
		}
	})
}
