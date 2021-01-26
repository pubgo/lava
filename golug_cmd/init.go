package golug_cmd

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/xerror"
)

func init() {
	// trace启动检查
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		xerror.Panic(dix_trace.Trigger())
	}))
}
