package debug

import (
	"net/http/pprof"

	"github.com/felixge/fgprof"
	"github.com/go-chi/chi/v5"

	"github.com/pubgo/lava/debug/debug_mux"
)

const Name = "debug"

func init() {
	debug_mux.DebugGet("/fgprof", fgprof.Handler().ServeHTTP)
}

func init() {
	debug_mux.Route(debug_mux.DebugPrefix("/pprof"), func(r chi.Router) {
		r.Get("/", pprof.Index)
		r.Get("/cmdline", pprof.Cmdline)
		r.Get("/profile", pprof.Profile)
		r.Get("/symbol", pprof.Symbol)
		r.Get("/trace", pprof.Trace)
		r.Get("/allocs", pprof.Handler("allocs").ServeHTTP)
		r.Get("/goroutine", pprof.Handler("goroutine").ServeHTTP)
		r.Get("/heap", pprof.Handler("heap").ServeHTTP)
		r.Get("/mutex", pprof.Handler("mutex").ServeHTTP)
		r.Get("/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
	})
}
