package tracelog

import (
	"expvar"
)

func Watch(name string, data func() interface{}) { expvar.Publish(name, expvar.Func(data)) }
