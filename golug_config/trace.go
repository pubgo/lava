package golug_config

import (
	"expvar"

	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish(Name, expvar.Func(func() interface{} {
			var data = make(map[string]interface{})
			for _, k := range GetCfg().AllKeys() {
				data[k] = GetCfg().GetString(k)
			}
			return data
		}))
	})
}
