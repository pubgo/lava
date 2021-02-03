package golug_broker

import (
	"github.com/pubgo/golug/golug_trace"
)

func init() {
	golug_trace.Watch(Name, func() interface{} {
		var data = make(map[string]string)
		for k, v := range List() {
			data[k] = v.Name()
		}
		return data
	})
}
