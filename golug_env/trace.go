package golug_env

import (
	"expvar"
	"os"
	"strings"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish("envs", expvar.Func(func() interface{} {
			var data []string
			for _, env := range os.Environ() {
				if strings.HasPrefix(env, Domain) {
					data = append(data, env)
				}
			}
			return data
		}))
	})
}
