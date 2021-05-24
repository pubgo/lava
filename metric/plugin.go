package metric

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/lug/watcher"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
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
			var dt map[string]Factory
			xerror.Panic(reporters.MapTo(&dt))
			return dt
		})
	},
	OnLog: func(logs xlog.Xlog) {
		_ = logs.Named(Name)
	},
}
