package metric_plugin

import (
	"context"
	"github.com/pubgo/lava/abc"
	"sync/atomic"
	"unsafe"

	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/logging/logkey"
	"github.com/pubgo/lava/core/metric"
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/vars"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: metric.Name,
		OnInit: func(p plugin.Process) {
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
			p.BeforeStop(func() { xerror.Panic(closer.Close()) })

			// 全局对象注册
			atomic.StorePointer(&g, unsafe.Pointer(&scope))
		},
		OnMiddleware: func(next abc.HandlerFunc) abc.HandlerFunc {
			return func(ctx context.Context, req abc.Request, resp func(rsp abc.Response) error) error {
				return next(metric.CreateCtx(ctx, GetGlobal()), req, resp)
			}
		},
		OnVars: func(v vars.Publisher) {
			v.Do(metric.Name+"_capabilities", func() interface{} {
				var c = GetGlobal().Capabilities()
				return typex.M{
					"reporting": c.Reporting(),
					"tagging":   c.Tagging(),
				}
			})

			v.Do(metric.Name+"_snapshot", func() interface{} {
				if c, ok := GetGlobal().(tally.TestScope); ok {
					// TODO 数据序列化处理
					return c.Snapshot()
				}
				return nil
			})
		},
	})
}
