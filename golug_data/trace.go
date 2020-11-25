package golug_data

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	xerror.Exit(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace {
			return
		}

		xlog.Debug("trace [data]")
		for k, v := range List() {
			fmt.Printf("%#v: \n\t%#v\n\n", k, v)
		}
		fmt.Println()
	}))
}
