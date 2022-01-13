package debug

import (
	"net/http/pprof"

	"github.com/felixge/fgprof"
	"github.com/go-chi/chi/v5"

	"github.com/pubgo/lava/mux"
)

const Name = "debug"

func init() {
	mux.Debug(func(r chi.Router) {
		r.Get("/fgprof", fgprof.Handler().ServeHTTP)
	})
}

func init() {
	mux.Debug(func(r chi.Router) {
		r.Route("/pprof", func(r chi.Router) {
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
	})
}
