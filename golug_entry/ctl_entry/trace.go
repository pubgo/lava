package ctl_entry

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func (t *ctlEntry) trace() {
	xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_config.Trace || !t.Options().Initialized {
			return
		}

		xlog.Debugf("ctl config trace")
		fmt.Println(golug_util.MarshalIndent(t.cfg))
		fmt.Println()
	}))
}
