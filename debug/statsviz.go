package debug

import (
	"net/http"

	"github.com/arl/statsviz"
	chiS "github.com/go-chi/chi/v5"
)

func init() {
	On(func(r *chiS.Mux) {
		r.Get("/debug/statsviz/ws", statsviz.Ws)
		r.Get("/debug/statsviz", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/debug/statsviz/", 301)
		})
		r.Handle("/debug/statsviz/*", statsviz.Index)
	})
}
