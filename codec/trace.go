package codec

import (
	"github.com/pubgo/golug/tracelog"
	"github.com/pubgo/xerror"
)

func init() {
	tracelog.Watch(Name, func() interface{} {
		var dt []string
		xerror.Panic(data.Each(func(key string) { dt = append(dt, key) }))
		return dt
	})
}
