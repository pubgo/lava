package golug_watcher

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
)

func init() {
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		for _, w := range List() {
			xerror.ExitF(w.Start(), w.String())
		}
	}))

	xerror.Exit(dix_run.WithBeforeStop(func(ctx *dix_run.BeforeStopCtx) {
		for _, w := range List() {
			xerror.ExitF(w.Close(), w.String())
		}
	}))
}
