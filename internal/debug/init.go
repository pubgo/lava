package debug

import (
	_ "expvar"
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
	http.HandleFunc("/debug/pprof/allocs", pprof.Handler("allocs").ServeHTTP)
	http.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	http.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	http.HandleFunc("/debug/pprof/mutex", pprof.Handler("mutex").ServeHTTP)
	http.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}

func init() {
	http.HandleFunc("/debug/statsviz/ws", statsviz.Ws)
	http.HandleFunc("/debug/statsviz/*", statsviz.Index)
	http.HandleFunc("/debug/statsviz", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/debug/statsviz/", 301)
	})
}
