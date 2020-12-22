package golug_db

import (
	"fmt"
	"os"

	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/xerror"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func(Name+"_cfg", func() interface{} { return cfg })

		var dbMetas = make(map[string][]*schemas.Table)
		ctx.Func(Name+"_dbMetas", func() interface{} { return dbMetas })
		clientM.Range(func(key, value interface{}) bool {
			engine := value.(*xorm.Engine)
			dbMetas[key.(string)] = xerror.PanicErr(engine.DBMetas()).([]*schemas.Table)
			engine.ShowSQL(true)
			xerror.Panic(engine.DumpAll(os.Stdout))
			return true
		})
		fmt.Println()
	})
}
