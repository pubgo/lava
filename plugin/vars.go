package plugin

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register(Name, func() interface{} {
		var data typex.Map
		for _, v := range All() {
			data.Set(v.ID(), v)
		}
		return data.Map()
	})

	vars.Register(Name+"_priority", func() interface{} {
		var data typex.A
		for _, key := range pluginKeys {
			data.Append(typex.Kv{Key: key.Value.(string), Value: key.Priority})
		}
		return data
	})
}
