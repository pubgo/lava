package golug_cmd

import (
	"github.com/pubgo/dix/dix_envs"
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/dix/dix_trace"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
)

func init() {
	// 运行环境检查
	xerror.Exit(dix_run.WithBeforeStart(func(ctx *dix_run.BeforeStartCtx) {
		var m = golug_env.RunMode
		switch golug_env.Mode {
		case m.Dev, m.Stag, m.Prod, m.Test, m.Release:
		default:
			xerror.Panic(xerror.Fmt("running mode does not match, mode: %s", golug_env.Mode))
		}
	}))

	// trace启动检查
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !dix_envs.IsTrace() {
			return
		}

		xerror.Next().Panic(dix_trace.Trace())
	}))
}
