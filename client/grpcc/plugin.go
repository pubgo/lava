package grpcc

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
			var cfg = GetDefaultCfg()
			if !config.Decode(Name, &cfg) {
				return
			}
		},
	})
}
