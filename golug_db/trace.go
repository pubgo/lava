package golug_db

import (
	"os"

	"github.com/pubgo/golug/golug_trace"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	golug_trace.Watch(Name+"_cfg", func() interface{} { return cfgMap })
	golug_trace.Watch(Name+"_dbMetas", func() interface{} {
		var dbMetas = make(map[string][]*schemas.Table)
		clientMap.Each(func(key, value interface{}) {
			engine := value.(*xorm.Engine)
			dbMetas[key.(string)] = xerror.PanicErr(engine.DBMetas()).([]*schemas.Table)
			engine.ShowSQL(true)
			xerror.Panic(engine.DumpAll(os.Stdout))
		})
		return dbMetas
	})
}
