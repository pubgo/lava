package metric

import (
	"context"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/runenv"
	"github.com/pubgo/lava/types"
)

var g = tally.NoopScope

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
				Tags:      Tags{"project": runenv.Project},
				Separator: cfg.Separator,
			}
			xerror.Exit(fc(config.GetMap(Name), &opts))

			scope, closer := tally.NewRootScope(opts, cfg.Interval)
			g = scope

			// 资源更新
			resource.Update("", &Resource{Scope: scope, Closer: closer})
		},
		OnMiddleware: func(next types.MiddleNext) types.MiddleNext {
			return func(ctx context.Context, req types.Request, resp func(rsp types.Response) error) error {

				ctx = ctxWith(ctx, g)

				return next(ctx, req, resp)
			}
		},
		OnVars: func(v types.Vars) {
			v.Do(Name+"_factory", func() interface{} {
				var dt = make(map[string]string)
				xerror.Panic(reporters.Each(func(name string, r Factory) {
					dt[name] = stack.Func(r)
				}))
				return dt
			})
		},
	})
}
