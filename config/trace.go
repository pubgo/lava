package config

import (
	"github.com/pubgo/golug/tracelog"
)

func init() {
	tracelog.Watch(Name, func() interface{} {
		var data = make(map[string]interface{})
		for _, k := range GetCfg().AllKeys() {
			data[k] = GetCfg().GetString(k)
		}
		return data
	})
}
