package db

import (
	"github.com/pubgo/golug/config"
	"github.com/pubgo/golug/plugin"
	"github.com/pubgo/golug/watcher"
	"github.com/pubgo/x/merge"
	"github.com/pubgo/xerror"
)

func init() { plugin.Register(plg) }

var plg = &plugin.Base{
	Name: Name,
	OnInit: func(ent interface{}) {
		if !config.Decode(Name, &cfgList) {
			return
		}

		for name := range cfgList {
			cfg := GetDefaultCfg()
			xerror.Panic(merge.Copy(&cfg, cfgList[name]))

			xerror.Panic(updateClient(name, *cfg))
			cfgList[name] = cfg
		}
	},
	OnWatch: func(name string, w *watcher.Response) {
		cfg, ok := cfgList[name]
		if !ok {
			cfg = GetDefaultCfg()
		}

		xerror.Panic(w.Decode(&cfg))
		xerror.Panic(updateClient(name, *cfg))
		cfgList[name] = cfg
	},
}
