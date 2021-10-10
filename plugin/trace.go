package plugin

import (
	"fmt"

	"github.com/pubgo/lug/pkg/typex"
	"github.com/pubgo/lug/vars"
)

func init() {
	vars.Watch(Name, func() interface{} {
		var data typex.Map
		for k, v := range All() {
			for i := range v {
				data.Set(fmt.Sprintf("%s.%s", k, v[i].Id()), v[i])
			}
		}
		return data.Map()
	})
}
