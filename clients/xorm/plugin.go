package xorm

import (
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"xorm.io/xorm/schemas"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/pkg/merge"
	"github.com/pubgo/lava/pkg/watcher"
	"github.com/pubgo/lava/plugin"
	"github.com/pubgo/lava/resource"
	"github.com/pubgo/lava/types"
)

func init() {
	plugin.Register(&plugin.Base{
		Name: Name,
		OnInit: func(p plugin.Process) {
			if config.Decode(Name, &cfgList) != nil {
				return
			}

			for name := range cfgList {
				cfgList[name] = merge.Copy(DefaultCfg(), cfgList[name]).(*Cfg)
				var db = cfgList[name].Build()
				resource.Update(name, &Client{Engine: db})

			}
		},
		OnWatch: func(name string, w *watcher.Response) (err error) {
			defer xerror.RespErr(&err)
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
			return nil
		},
		OnVars: func(v types.Vars) {
			v(Name+"_cfg", func() interface{} { return cfgList })
			v(Name+"_dbMetas", func() interface{} {
				var dbMetas = make(map[string][]*schemas.Table)
				for name, res := range resource.GetByKind(Name) {
					dbMetas[name] = xerror.PanicErr(res.(*Client).DBMetas()).([]*schemas.Table)
				}
				return dbMetas
			})

			v(Name+"_sqlList", func() interface{} {
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
