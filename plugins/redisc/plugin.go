package redisc

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/entry"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

func init() { plugin.Register(&plg) }

var plg = plugin.Base{
	Name: Name,
	OnInit: func(ent entry.Entry) {
		config.Decode(Name, &cfg)

		for k, v := range cfg {
			cfg1 := GetDefaultCfg()
			xerror.Panic(merge.Copy(&cfg1, v))
			Update(k, cfg1)
			cfg[k] = cfg1
		}
	},
}
