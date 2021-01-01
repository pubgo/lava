package grpclient

import (
	"github.com/pubgo/dix/dix_trace"
)

func init() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func(Name+"_cfg", func() interface{} { return cfg })
		ctx.Func(Name+"_interceptor", func() interface{} {
			var interceptors []string
			interceptorMap.Range(func(key, value interface{}) bool {
				interceptors = append(interceptors, key.(string))
				return true
			})
			return interceptors
		})
	})
}
