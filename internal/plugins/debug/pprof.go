package debug

import (
	_ "expvar"
	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lava/mux"
	"net/http"
	"net/http/pprof"

	"github.com/arl/statsviz"
	"github.com/felixge/fgprof"
	_ "golang.org/x/net/trace"
)

const Name = "debug"

func init() {
	http.HandleFunc("/debug/fgprof", fgprof.Handler().ServeHTTP)
}

func init() {
	mux.Route("/debug/pprof", func(r chi.Router) {
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

func init() {
	mux.Route("/debug/statsviz", func(r chi.Router) {
		r.Get("/ws", statsviz.Ws)
		r.Get("/", statsviz.Index)
		r.Get("/", statsviz.Index)
	})
}
