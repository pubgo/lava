package plugin

import (
	"github.com/pubgo/golug/tracelog"
)

func init() {
	tracelog.Watch("plugin", func() interface{} {
		var data = make(map[string][]string)
		for k, v := range All() {
			for i := range v {
				data[k] = append(data[k], v[i].String())
			}
		}
		return data
	})
}
