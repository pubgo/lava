package golug_version

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/pkg/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace {
			return
		}

		xlog.Debug("trace [version]")
		for name, v := range List() {
			fmt.Println(name, golug_util.MarshalIndent(v))
		}
		fmt.Println()
	}))
}
