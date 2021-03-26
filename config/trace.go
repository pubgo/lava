package config

import (
	"github.com/pubgo/golug/vars"
)

func init() {
	vars.Watch("config", func() interface{} {
		var data = make(map[string]interface{})
		for _, k := range GetCfg().AllKeys() {
			data[k] = GetCfg().GetString(k)
		}
		return data
	})
}
