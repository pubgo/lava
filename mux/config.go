package mux

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"net/http"
	"time"
)

const Name = "mux"

var addr = ":8088"
var app = chi.NewRouter()
var server = &http.Server{Addr: addr, Handler: app}

func Default() *chi.Mux { return app }

func init() {
	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)
	app.Use(middleware.RequestID)
	app.Use(middleware.Timeout(60 * time.Second))
}
