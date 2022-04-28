package metric_builder

import (
	"context"
	"sync/atomic"
	"unsafe"

	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
)

func Builder(srv service.Service) {
	var cfg = metric.DefaultCfg()
	_ = config.Decode(metric.Name, &cfg)

	driver := cfg.Driver
	xerror.Assert(driver == "", "metric driver is null")

	fc := metric.GetFactory(driver)
	xerror.Assert(fc == nil, "metric driver [%s] not found", driver)

	var opts = tally.ScopeOptions{
		Tags:      metric.Tags{logkey.Project: runtime.Name()},
		Separator: cfg.Separator,
	}
	xerror.Exit(fc(config.GetMap(metric.Name), &opts))

	scope, closer := tally.NewRootScope(opts, cfg.Interval)
	srv.BeforeStops(func() { xerror.Panic(closer.Close()) })

	// 全局对象注册
	atomic.StorePointer(&g, unsafe.Pointer(&scope))
}

func init() {
	middleware.Register(metric.Name, func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx context.Context, req middleware.Request, resp middleware.Response) error {
			return next(metric.CreateCtx(ctx, GetGlobal()), req, resp)
		}
	})
}
