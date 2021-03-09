package mux

import (
	"expvar"
	"fmt"
)

func init() {
	expvar.Publish(Name+"_rest_router", expvar.Func(func() interface{} {
		if app == nil {
			return nil
		}

		var data []string
		for _, stack := range app.Routes() {
			data = append(data, fmt.Sprintf("%#v\n", stack))
		}
		return data
	}))
}
