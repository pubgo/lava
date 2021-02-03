package golug_rest

import (
	"github.com/pubgo/golug/golug_trace"
)

func (t *restEntry) trace() {
	golug_trace.Watch(t.Options().Name+"_rest_router", func() interface{} {
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
}
