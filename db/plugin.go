package db

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/gutils"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent interface{}) {
			config.Decode(Name, &cfgMap)

			for k, v := range cfgMap {
				cfg := GetDefaultCfg()
				gutils.Mergo(&cfg, v)

				initClient(k, cfg)
				cfgMap[k] = cfg
			}
		},
	})
}
