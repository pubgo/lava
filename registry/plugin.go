package registry

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/xerror"
)

func init() { plugin.Register(&plg) }

var plg = plugin.Base{
	Name: Name,
	OnInit: func(ent entry.Entry) {
		var cfg = GetDefaultCfg()
		if !config.Decode(Name, &cfg) {
			return
		}

		defaultRegistry = xerror.PanicErr(cfg.Build()).(Registry)
	},
}
