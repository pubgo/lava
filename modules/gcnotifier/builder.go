package gcnotifier

import (
	"context"

	"github.com/CAFxX/gcnotifier"
	"github.com/pubgo/funk/async"
	"github.com/pubgo/funk/log"
	"github.com/pubgo/funk/running"
	"go.uber.org/atomic"

	"github.com/pubgo/lava/core/lifecycle"
)

var Name = "gc"

func New(log log.Logger) lifecycle.Handler {
	if !running.IsDebug {
		return nil
	}

	logs := log.WithName(Name)
	var num atomic.Uint64
	return func(lc lifecycle.Lifecycle) {
		lc.AfterStart(func() {
			lc.BeforeStop(async.GoCtx(func(ctx context.Context) error {
				gc := gcnotifier.New()
				defer gc.Close()

				// TODO handler

				for {
					select {
					case <-gc.AfterGC():
						num.Add(1)
						logs.Info().Uint64("num", num.Load()).Msg("gc notify")
					case <-ctx.Done():
						return nil
					}
				}
			}))
		})
	}
}
