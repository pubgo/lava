package golug_db

import (
	"fmt"
	"os"

	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		xlog.Debugf("%s client trace", Name)
		fmt.Println(golug_util.MarshalIndent(cfg))
		clientM.Range(func(key, value interface{}) bool {
			engine := value.(*xorm.Engine)
			fmt.Println("DBMetas", key.(string), golug_util.MarshalIndent(xerror.PanicErr(engine.DBMetas()).([]*schemas.Table)))
			engine.ShowSQL(true)
			xerror.Panic(engine.DumpAll(os.Stdout))
			return true
		})
		fmt.Println()
	})
}
