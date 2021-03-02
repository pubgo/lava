package broker

import (
	"github.com/pubgo/golug/tracelog"
)

func init() {
	tracelog.Watch(Name, func() interface{} {
		var data = make(map[string]string)
		for k, v := range List() {
			data[k] = v.Name()
		}
		return data
	})
}
