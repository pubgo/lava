package metric

import (
	"time"

	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
	"github.com/uber-go/tally"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/runenv"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			var cfg = GetDefaultCfg()
			_ = config.Decode(Name, &cfg)

			driver := cfg.Driver
			xerror.Assert(driver == "", "metric driver is null")

			fc := Get(driver)
			xerror.Assert(fc == nil, "metric driver %s not found", driver)

			var opts = tally.ScopeOptions{Prefix: runenv.Project}
			xerror.Exit(fc(config.GetMap(Name), &opts))
			scope, closer := tally.NewRootScope(opts, time.Second)
			ent.AfterStop(func() { xerror.Panic(closer.Close()) })
			setDefault(scope)
		},
		OnVars: func(w func(name string, data func() interface{})) {
			w(Name, func() interface{} {
				var dt = make(map[string]string)
				xerror.Panic(reporters.Each(func(name string, r Factory) {
					dt[name] = stack.Func(r)
				}))
				return dt
			})
		},
	})
}
