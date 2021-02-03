package golug_codec

import (
	"github.com/pubgo/golug/golug_trace"
)

func init() {
	golug_trace.Watch(Name, func() interface{} {
		var dt []string
		data.Each(func(key string) { dt = append(dt, key) })
		return dt
	})
}
