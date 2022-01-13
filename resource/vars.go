package resource

import (
	"fmt"

	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register(Name, func() interface{} {
		var data = make(map[string]map[string]string)
		sources.Range(func(name string, val interface{}) bool {
			var kind = val.(Resource).Kind()
			if data[kind] == nil {
				data[kind] = make(map[string]string)
			}
			data[kind][name] = fmt.Sprintf("%#v", val)
			return true
		})
		return data
	})
}
