package db

import (
	"github.com/pubgo/golug/vars"
	"github.com/pubgo/x/strutil"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	vars.Watch(Name+"_cfg", func() interface{} { return cfgList })
	vars.Watch(Name+"_dbMetas", func() interface{} {
		var dbMetas = make(map[string][]*schemas.Table)
		xerror.Panic(clients.Each(func(key string, engine *xorm.Engine) {
			dbMetas[key] = xerror.PanicErr(engine.DBMetas()).([]*schemas.Table)
		}))
		return dbMetas
	})

	vars.Watch(Name+"_sqlList", func() interface{} {
		var sqlList []string
		xerror.Panic(clients.Each(func(key string, engine *xorm.Engine) {
			var b strutil.Builder
			defer b.Reset()
			xerror.Panic(engine.DumpAll(&b))
			sqlList = append(sqlList, b.String())
		}))
		return sqlList
	})
}
