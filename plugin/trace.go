package plugin

import (
	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch(Name, func() interface{} {
		var data typex.Map
		for k, v := range All() {
			for i := range v {
				data.Set(k, v[i])
			}
		}
		return data.Map()
	})
}
