package db

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/xerror"
)

func init() {
	var onInit = func(ent interface{}) {
		config.Decode(Name, &cfgList)

		for k, v := range cfgList {
			cfg := GetDefaultCfg()
			xerror.Panic(gutils.Mergo(&cfg, v))

			xerror.Panic(initClient(k, cfg))
			cfgList[k] = cfg
		}
	}

	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
