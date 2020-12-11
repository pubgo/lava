package golug_task

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func (t *taskEntry) trace() {
	xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace || !t.Options().Initialized {
			return
		}

		xlog.Debugf("task config trace")
		fmt.Println(golug_util.MarshalIndent(t.cfg))
		fmt.Println()
	}))
}
