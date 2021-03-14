package db

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/gutils"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/watcher"
	"github.com/pubgo/xerror"
)

func onInit(ent interface{}) {
	if !config.Decode(Name, &cfgList) {
		return
	}

	for name := range cfgList {
		cfg := GetDefaultCfg()
		xerror.Panic(gutils.Mergo(&cfg, cfgList[name]))

		xerror.Panic(updateClient(name, cfg))
		cfgList[name] = cfg
	}
}

func init() {
	plugin.Register(&plugin.Base{
		Name:   Name,
		OnInit: onInit,
		OnWatch: func(name string, w *watcher.Response) {
			cfg, ok := cfgList[name]
			if !ok {
				cfg = GetDefaultCfg()
			}

			xerror.Panic(w.Decode(&cfg))

			xerror.Panic(updateClient(name, cfg))
			cfgList[name] = cfg
		},
	})
}
