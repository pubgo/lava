package golug_env

import (
	"fmt"
	"os"
	"strings"

	"github.com/pubgo/dix/dix_run"
	"github.com/pubgo/xerror"
	"github.com/pubgo/xlog"
)

func init() {
	xerror.Panic(dix_run.WithAfterStart(func(ctx *dix_run.AfterStartCtx) {
		if !Trace {
			return
		}

		xlog.Debug("trace [env]")
		for _, env := range os.Environ() {
			if strings.HasPrefix(env, Domain) {
				fmt.Println(env)
			}
		}
		fmt.Println()
	}))
}
