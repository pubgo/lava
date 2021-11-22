package plugin

import (
	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register(Name, func() interface{} {
		var data typex.Map
		for _, v := range All() {
			data.Set(v.UniqueName(), v)
		}
		return data.Map()
	})
}
