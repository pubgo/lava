package golug_db

import (
	"fmt"
	"os"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/pkg/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
	"xorm.io/xorm"
	"xorm.io/xorm/schemas"
)

func init() {
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace {
			return
		}

		xlog.Debugf("%s client trace", name)
		fmt.Println(golug_util.MarshalIndent(cfg))
		clientM.Range(func(key, value interface{}) bool {
			engine := value.(*xorm.Engine)
			fmt.Println("DBMetas",key.(string), golug_util.MarshalIndent(xerror.PanicErr(engine.DBMetas()).([]*schemas.Table)))
			engine.ShowSQL(true)
			xerror.Panic(engine.DumpAll(os.Stdout))
			return true
		})

		fmt.Println()
	}))
}
