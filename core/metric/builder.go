package metric

import (
	"context"
	"sync/atomic"
	"unsafe"

	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/lifecycle"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging/logkey"
	"github.com/pubgo/lava/middleware"
	"github.com/pubgo/lava/runtime"
)

func init() {
	inject.Provide(func() *Cfg {
		var cfg = DefaultCfg()
		_ = config.Decode(Name, &cfg)
		xerror.Panic(config.UnmarshalKey(Name, &cfg))
		driver := cfg.Driver
		xerror.Assert(driver == "", "metric driver is null")
		return &cfg
	})

	inject.Invoke(
		func(in struct {
			M    lifecycle.Lifecycle
			Cfg  *Cfg
			Opts []*tally.ScopeOptions `group:"metric"`
		}) {

		},
	)

	middleware.Register(Name, func(next middleware.HandlerFunc) middleware.HandlerFunc {
		return func(ctx context.Context, req middleware.Request, resp middleware.Response) error {
			return next(CreateCtx(ctx, GetGlobal()), req, resp)
		}
	})
}

func Builder(m lifecycle.Lifecycle) {
	var cfg = DefaultCfg()
	_ = config.Decode(Name, &cfg)

	driver := cfg.Driver
	xerror.Assert(driver == "", "metric driver is null")

	fc := GetFactory(driver)
	xerror.Assert(fc == nil, "metric driver [%s] not found", driver)

	var opts = tally.ScopeOptions{
		Tags:      Tags{logkey.Project: runtime.Project},
		Separator: cfg.Separator,
	}
	xerror.Exit(fc(config.GetMap(Name), &opts))

	scope, closer := tally.NewRootScope(opts, cfg.Interval)
	m.BeforeStops(func() { xerror.Panic(closer.Close()) })

	// 全局对象注册
	atomic.StorePointer(&g, unsafe.Pointer(&scope))
}
