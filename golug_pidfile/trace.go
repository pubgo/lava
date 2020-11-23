package golug_pidfile

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace {
			return
		}

		pid, err := GetPid()
		xerror.Panic(err)
		xlog.Debug("pidfile trace", xlog.Int("pid", pid), xlog.String("path", GetPidPath()))
		fmt.Println()
	}))
}
