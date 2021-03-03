package config

import (
	"github.com/pubgo/golug/tracelog"
)

func init() {
	tracelog.Watch(name, func() interface{} {
		var data = make(map[string]interface{})
		for _, k := range GetCfg().AllKeys() {
			data[k] = GetCfg().GetString(k)
		}
		return data
	})
}
