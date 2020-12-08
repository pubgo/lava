package golug_watcher

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/golug_util"
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
		fmt.Println(golug_util.MarshalIndent(List()))
		fmt.Println()
	}))
}
