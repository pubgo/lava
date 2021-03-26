package codec

import (
	"github.com/pubgo/golug/vars"
	"github.com/pubgo/xerror"
)

func init() {
	vars.Watch(Name, func() interface{} {
		var dt []string
		xerror.Panic(data.Each(func(key string) { dt = append(dt, key) }))
		return dt
	})
}
