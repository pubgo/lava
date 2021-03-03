package db

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/xerror"
)

func init() {
	var onInit = func(ent interface{}) {
		config.Decode(Name, &cfgMap)

		for k, v := range cfgMap {
			cfg := GetDefaultCfg()
			xerror.Panic(gutils.Mergo(&cfg, v))

			initClient(k, cfg)
			cfgMap[k] = cfg
		}
	}
	
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
