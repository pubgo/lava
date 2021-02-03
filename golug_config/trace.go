package golug_config

import (
	"github.com/pubgo/golug/golug_trace"
)

func init() {
	golug_trace.Watch(Name, func() interface{} {
		var data = make(map[string]interface{})
		for _, k := range GetCfg().AllKeys() {
			data[k] = GetCfg().GetString(k)
		}
		return data
	})
}
