package codec

import (
	"github.com/pubgo/golug/tracelog"
)

func init() {
	tracelog.Watch(Name, func() interface{} {
		var dt []string
		data.Each(func(key string) { dt = append(dt, key) })
		return dt
	})
}
