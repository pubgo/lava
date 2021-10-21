package registry

import (
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			var cfg = GetDefaultCfg()
			if config.Decode(Name, &cfg) != nil {
				return
			}

			defaultRegistry = xerror.PanicErr(cfg.Build()).(Registry)
		},
	})
}
