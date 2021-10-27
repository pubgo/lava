package debug

import (
	"github.com/arl/statsviz"
	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lava/mux"
)

func init() {
	mux.Route("/debug/statsviz", func(r chi.Router) {
		r.Get("/ws", statsviz.Ws)
		r.Get("/", statsviz.Index)
		r.Get("/", statsviz.Index)
	})
}
