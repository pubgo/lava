package logging

import "github.com/pubgo/lava/vars"

func init() {
	vars.Register(name, func() interface{} {
		var keys []string
		loggerMap.Range(func(key, _ interface{}) bool {
			keys = append(keys, key.(string))
			return true
		})
		return keys
	})
}
