package mux

import (
	"expvar"
)

func init() {
	Default().Handle("/debug/vars", expvar.Handler())
}
