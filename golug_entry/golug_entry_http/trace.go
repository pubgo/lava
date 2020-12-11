package golug_entry_http

import (
	"fmt"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/golug/golug_env"
	"github.com/pubgo/golug/internal/golug_util"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func (t *httpEntry) trace() {
	xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !golug_env.Trace || !t.Options().Initialized {
			return
		}

		xlog.Debug("http rest router trace")
		for _, stacks := range t.app.Stack() {
			for _, stack := range stacks {
				if stack.Path == "/" {
					continue
				}

				log.Debugf("%s %s", stack.Method, stack.Path)
			}
		}
		fmt.Println()

		xlog.Debugf("http server config trace")
		fmt.Println(golug_util.MarshalIndent(t.cfg))
		fmt.Println()
	}))
}
