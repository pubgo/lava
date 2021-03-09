package tracelog

import (
	"expvar"
	
	"github.com/pubgo/golug/mux"
)

func init() {
	mux.Default().Handle("/debug/vars", expvar.Handler())
}
