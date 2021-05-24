package redisc

import (
	"github.com/pubgo/lug/config"
	"github.com/pubgo/lug/plugin"
	"github.com/pubgo/x/merge"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
			config.Decode(Name, &cfg)

			for k, v := range cfg {
				_cfg := GetDefaultCfg()
				merge.Copy(&_cfg, v)
				initClient(k, _cfg)
				cfg[k] = _cfg
			}
		},
	})
}
