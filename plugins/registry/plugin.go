package registry

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			var cfg = GetDefaultCfg()
			if !config.Decode(Name, &cfg) {
				return
			}

			defaultRegistry = xerror.PanicErr(cfg.Build()).(Registry)
		},
	})
}
