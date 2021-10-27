package plugin

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Watch(Name, func() interface{} {
		var data typex.Map
		for k, v := range All() {
			data.Set(k, v)
		}
		return data.Map()
	})
}
