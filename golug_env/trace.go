package golug_env

import (
	"os"
	"strings"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.TraceCtx) {
		ctx.Func("envs", func() interface{} {
			var data []string
			for _, env := range os.Environ() {
				if strings.HasPrefix(env, Prefix) {
					data = append(data, env)
				}
			}
			return data
		})
	})
}
