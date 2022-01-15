package metric

import (
	"context"

	"github.com/pubgo/dix"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logger/logkey"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runtime"
	"github.com/pubgo/lava/types"
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
			g.Store(scope)

			// 注入依赖scope
			xerror.Panic(dix.Provider(scope))
		},
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {
				return next(CreateCtxWith(ctx, GetGlobal()), req, resp)
			}
		},
		OnVars: func(v types.Vars) {
			v.Do(Name+"_factory", func() interface{} {
				var dt = make(map[string]string)
				xerror.Panic(factories.Each(func(name string, r Factory) {
					dt[name] = stack.Func(r)
				}))
				return dt
			})

			v.Do(Name+"_capabilities", func() interface{} {
				var c = GetGlobal().Capabilities()
				return types.M{
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
