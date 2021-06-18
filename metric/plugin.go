package metric

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/x/stack"
	"github.com/pubgo/xerror"
)

func init() { plugin.Register(&plg) }

var plg = plugin.Base{
	Name: Name,
	OnInit: func(ent interface{}) {
		var cfg = GetDefaultCfg()
		if !config.Decode(Name, &cfg) {
			return
		}

		var reporter = xerror.PanicErr(cfg.Build()).(Reporter)
		setDefault(reporter)
	},
	OnWatch: func(name string, resp *watcher.Response) {
		var cfg = GetDefaultCfg()
		_ = config.Decode(Name, &cfg)
		xerror.Panic(resp.Decode(&cfg))

		var reporter = xerror.PanicErr(cfg.Build()).(Reporter)
		setDefault(reporter)
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
}
