package pidfile

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/entry"
	"github.com/pubgo/lava/plugin"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnInit: func(ent entry.Entry) {
		var cfg Cfg
		_ = config.Decode(Name, &cfg)

		if cfg.PidPath != "" {
			pidPath = cfg.PidPath
		}
	},
}
