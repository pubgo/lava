package metric

import (
	"context"
	"sync/atomic"
	"unsafe"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/service"
	"github.com/pubgo/lava/vars"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			var cfg = DefaultCfg()
			_ = config.Decode(Name, &cfg)

			driver := cfg.Driver
			xerror.Assert(driver == "", "metric driver is null")

			fc := GetFactory(driver)
			xerror.Assert(fc == nil, "metric driver [%s] not found", driver)

			var opts = tally.ScopeOptions{
				Tags:      Tags{logkey.Project: runtime.Name()},
				Separator: cfg.Separator,
			}
			xerror.Exit(fc(config.GetMap(Name), &opts))

			scope, closer := tally.NewRootScope(opts, cfg.Interval)
			p.BeforeStop(func() { xerror.Panic(closer.Close()) })

			// 全局对象注册
			atomic.StorePointer(&g, unsafe.Pointer(&scope))
		},
		OnMiddleware: func(next service.HandlerFunc) service.HandlerFunc {
			return func(ctx context.Context, req service.Request, resp func(rsp service.Response) error) error {
				return next(CreateCtx(ctx, GetGlobal()), req, resp)
			}
		},
		OnVars: func(v vars.Publisher) {
			v.Do(Name+"_factory", func() interface{} {
				var dt = make(map[string]string)
				xerror.Panic(factories.Each(func(name string, r Factory) {
					dt[name] = stack.Func(r)
				}))
				return dt
			})

			v.Do(Name+"_capabilities", func() interface{} {
				var c = GetGlobal().Capabilities()
				return typex.M{
					"reporting": c.Reporting(),
					"tagging":   c.Tagging(),
				}
			})

			v.Do(Name+"_snapshot", func() interface{} {
				if c, ok := GetGlobal().(tally.TestScope); ok {
					// TODO 数据序列化处理
					return c.Snapshot()
				}
				return nil
			})
		},
	})
}
