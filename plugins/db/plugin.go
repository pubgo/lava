package db

import (
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"xorm.io/xorm/schemas"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/types"
	"github.com/pubgo/lava/watcher"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(ent plugin.Entry) {
			if config.Decode(Name, &cfgList) != nil {
				return
			}

			for name := range cfgList {
				cfgList[name] = merge.Copy(DefaultCfg(), cfgList[name]).(*Cfg)
				var db = cfgList[name].Build()
				resource.Update(name, &Client{Engine: db})

			}
		},
		OnWatch: func(name string, w *watcher.Response) {
			w.OnPut(func() {
				cfg, ok := cfgList[name]
				if !ok {
					cfg = DefaultCfg()
				}
				xerror.Panic(types.Decode(w.Value, &cfg))
				cfgList[name] = cfg

				var db = cfgList[name].Build()
				resource.Update(name, &Client{Engine: db})
			})

			w.OnDelete(func() { resource.Remove(Name, name) })
		},
		OnVars: func(w func(name string, data func() interface{})) {
			w(Name+"_cfg", func() interface{} { return cfgList })
			w(Name+"_dbMetas", func() interface{} {
				var dbMetas = make(map[string][]*schemas.Table)
				for name, res := range resource.GetByKind(Name) {
					dbMetas[name] = xerror.PanicErr(res.(*Client).DBMetas()).([]*schemas.Table)
				}
				return dbMetas
			})

			w(Name+"_sqlList", func() interface{} {
				var sqlList = make(map[string]string)
				for name, res := range resource.GetByKind(Name) {
					var b strutil.Builder
					xerror.Panic(res.(*Client).DumpAll(&b))
					b.Reset()
					sqlList[name] = b.String()
				}
				return sqlList
			})
		},
	})
}
