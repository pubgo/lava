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

		return fmt.Sprintf("%#v\n", app.Routes())
	}))
}
