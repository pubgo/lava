package healthy

import (
	"context"
	"time"

	"github.com/pubgo/x/fx"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/metric"
	"github.com/pubgo/lug/pkg/ctxutil"
)

func init() {
	entry.AfterStart(func() {
		var healthCheckTimer = metric.NewTimer("health_check")
		var scope = metric.WithSubScope("health_check")
		entry.AfterStop(fx.Tick(func(ctx fx.Ctx) {
			metric.TimeRecord(healthCheckTimer, func() {
				for name, r := range healthList.Map() {
					metric.TimeRecord(scope.Timer(name), func() {
						var ctx, cancel = context.WithTimeout(ctx, ctxutil.DefaultTimeout)
						defer cancel()
						xerror.Panic(r.(HealthCheck)(ctx))
					})
				}
			})
		}, time.Second))
	})
}
