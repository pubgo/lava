package golug_version

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_config"
	"github.com/pubgo/golug/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_config.Trace {
			return
		}

		for name, v := range List() {
			xlog.Debug(name)
			fmt.Println(golug_util.MarshalIndent(v))
		}
		fmt.Println()
	}))
}
