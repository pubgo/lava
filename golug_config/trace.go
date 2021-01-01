package golug_config

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func(Name, func() interface{} {
			var data = make(map[string]interface{})
			for _, k := range GetCfg().AllKeys() {
				data[k] = GetCfg().GetString(k)
			}
			return data
		})
	})
}
