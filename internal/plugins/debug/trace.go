package debug

import (
	"github.com/go-chi/chi/v5"
	"golang.org/x/net/trace"

	"github.com/pubgo/lava/mux"
)

func init() {
	mux.Debug(func(r chi.Router) {
		r.Get("/requests", trace.Traces)
		r.Get("/events", trace.Events)
	})
}
