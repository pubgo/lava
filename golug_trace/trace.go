package golug_trace

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
)

type LogCtx struct{ dix.Model }

func Log(fn func(_ *LogCtx)) {
	xerror.Next().Panic(dix.Dix(fn))
}

func init() {
	// 服务启动之后, 然后打印trace信息
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace {
			return
		}

		xerror.Panic(dix.Dix(LogCtx{}))
	}))
}
