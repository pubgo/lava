package debug

import (
	"github.com/go-chi/chi/v5"

	"net/http/pprof"
)

func init() { On(profRoute) }

func profRoute(app *chi.Mux) {
	app.HandleFunc("/debug/pprof", pprof.Index)
	app.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	app.HandleFunc("/debug/pprof/profile", pprof.Profile)
	app.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	app.HandleFunc("/debug/pprof/trace", pprof.Trace)
	app.HandleFunc("/debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
	app.HandleFunc("/debug/pprof/block", pprof.Handler("block").ServeHTTP)
	app.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	app.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	app.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	app.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}
