package golug_rest

import (
	"expvar"
	"github.com/pubgo/dix/dix_trace"
)

func (t *restEntry) trace() {
	dix_trace.With(func(_ *dix_trace.TraceCtx) {
		expvar.Publish(t.Options().Name+"_rest_router", expvar.Func(func() interface{} {
			var data []map[string]string
			if t.app == nil {
				return nil
			}

			for i, stacks := range t.app.Stack() {
				data = append(data, make(map[string]string))
				for _, stack := range stacks {
					if stack == nil {
						continue
					}

					if stack.Path == "/" {
						continue
					}
					data[i][stack.Method] = stack.Path
				}
			}
			return data
		}))
	})
}
