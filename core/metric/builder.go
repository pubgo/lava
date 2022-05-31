package metric

import (
	"context"
	"sync/atomic"
	"unsafe"

	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/runtime"
)

func init() {
	dix.Register(func() *Cfg {
		var cfg = DefaultCfg()
		_ = config.Decode(Name, &cfg)
		xerror.Panic(config.UnmarshalKey(Name, &cfg))
		driver := cfg.Driver
		xerror.Assert(driver == "", "metric driver is null")
		return &cfg
	})

	dix.Register(func(m lifecycle.Lifecycle, cfg *Cfg, sopts map[string]*tally.ScopeOptions) Metric {
		driver := cfg.Driver
		xerror.Assert(driver == "", "metric driver is null")

		var opts = tally.ScopeOptions{
			Tags:      Tags{logkey.Project: runtime.Project},
			Separator: cfg.Separator,
		}

		_ = sopts
		scope, closer := tally.NewRootScope(opts, cfg.Interval)
		m.BeforeStops(func() { xerror.Panic(closer.Close()) })

		// 全局对象注册
		atomic.StorePointer(&g, unsafe.Pointer(&scope))
		return scope
	})

	dix.Register(func(m Metric) {
		middleware.Register(Name, func(next middleware.HandlerFunc) middleware.HandlerFunc {
			return func(ctx context.Context, req middleware.Request, resp middleware.Response) error {
				return next(CreateCtx(ctx, m), req, resp)
			}
		})
	})
}
