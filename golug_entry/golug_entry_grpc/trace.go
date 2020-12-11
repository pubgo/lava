package golug_entry_grpc

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
)

func (t *grpcEntry) trace() {
	xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace || !t.Options().Initialized {
			return
		}

	}))
}
