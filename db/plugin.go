package db

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	config.Decode(Name, &cfgList)

	for name := range cfgList {
		cfg := GetDefaultCfg()
		xerror.Panic(gutils.Mergo(&cfg, cfgList[name]))

		xerror.Panic(initClient(name, cfg))
		cfgList[name] = cfg
	}
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
	})
}
