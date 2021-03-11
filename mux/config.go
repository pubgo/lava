package mux

import (
	"github.com/go-chi/chi/v5"

	"net/http"
)

const Name = "mux"

var addr = ":8088"
var app = chi.NewMux()
var server = &http.Server{Addr: addr, Handler: app}

func Default() *chi.Mux { return app }
