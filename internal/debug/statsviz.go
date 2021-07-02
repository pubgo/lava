package debug

import (
	"net/http"

	"github.com/arl/statsviz"
	"github.com/go-chi/chi/v5"
)

func init() {
	On(func(r *chi.Mux) {
		r.Get("/debug/statsviz/ws", statsviz.Ws)
		r.Handle("/debug/statsviz/*", statsviz.Index)
		r.Get("/debug/statsviz", func(w http.ResponseWriter, r *http.Request) {
			http.Redirect(w, r, "/debug/statsviz/", 301)
		})
	})
}
