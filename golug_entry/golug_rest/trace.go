package golug_rest

import (
	"github.com/pubgo/dix/dix_trace"
)

func (t *restEntry) trace() {
	dix_trace.With(func(ctx *dix_trace.Ctx) {
		ctx.Func(t.Options().Name+"_rest_router", func() interface{} {
			if t.app == nil {
				return nil
			}

			var data []map[string]string
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
		})
	})
}
