package rest

import (
	"github.com/pubgo/lug/vars"
)

func (t *restEntry) trace() {
	vars.Watch(t.Options().Name+"_cfg", func() interface{} { return t.cfg })
	vars.Watch(t.Options().Name+"_rest_router", func() interface{} {
		if t.srv.Get() == nil {
			return nil
		}

		var data []map[string]string
		for i, stacks := range t.srv.Get().Stack() {
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
