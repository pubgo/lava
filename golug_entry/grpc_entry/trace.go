package grpc_entry

import (
	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/xerror"
)

func (t *grpcEntry) trace() {
	xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_config.Trace || !t.opts.Initialized {
			return
		}

	}))
}
