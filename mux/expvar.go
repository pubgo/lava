package mux

import (
	"expvar"
)

func init() {
	Default().Handle("/", expvar.Handler())
}
