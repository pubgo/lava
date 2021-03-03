package db

import (
	"os"

	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	tracelog.Watch(Name+"_default_cfg", func() interface{} { return GetDefaultCfg() })
	tracelog.Watch(Name+"_cfg", func() interface{} { return cfgMap })
	tracelog.Watch(Name+"_dbMetas", func() interface{} {
		var dbMetas = make(map[string][]*schemas.Table)
		clientMap.Each(func(key string, engine *xorm.Engine) {
			dbMetas[key] = xerror.PanicErr(engine.DBMetas()).([]*schemas.Table)
			engine.ShowSQL(true)
			xerror.Panic(engine.DumpAll(os.Stdout))
		})
		return dbMetas
	})
}
