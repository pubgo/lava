package golug_watcher

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	// debug and trace
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace {
			return
		}

		xlog.Debug("trace [log] config")
		var dt []string
		dataCallback.Range(func(key, _ interface{}) bool { dt = append(dt, key.(string)); return true })
		fmt.Println(golug_util.MarshalIndent(dt))
		fmt.Println()
	}))
}
