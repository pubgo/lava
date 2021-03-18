package mux

import (
	"expvar"

	"github.com/go-chi/chi/v5"
)

func init() {
	On(func(app *chi.Mux) {
		app.Handle("/", expvar.Handler())
	})
}
