package resource

import (
	"fmt"

	"github.com/pubgo/lava/pkg/typex"
	"github.com/pubgo/lava/vars"
)

func init() {
	vars.Register(Name, func() interface{} {
		var data = make(map[string][]typex.Kv)
		resourceList.Range(func(name string, val interface{}) bool {
			var kind = val.(Resource).Kind()
			data[kind] = append(data[kind], typex.KvOf(name, fmt.Sprintf("%#v", val)))
			return true
		})
		return data
	})
}
