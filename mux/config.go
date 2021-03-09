package mux

import (
	"github.com/go-chi/chi/v5"
)

const Name = "mux"

var addr = ":8088"
var app = chi.NewMux()

func Default() *chi.Mux { return app }
