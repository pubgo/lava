package pidfile

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
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
