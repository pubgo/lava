package pidfile

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			var cfg Cfg
			_ = config.Decode(Name, &cfg)

			if cfg.PidPath != "" {
				pidPath = cfg.PidPath
			}
		},
	})
}
