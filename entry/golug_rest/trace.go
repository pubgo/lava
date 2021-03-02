package golug_rest

import (
	"github.com/pubgo/golug/tracelog"
)

func (t *restEntry) trace() {
	tracelog.Watch(t.Options().Name+"_rest_router", func() interface{} {
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
