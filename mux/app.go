package mux

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maruel/panicparse/v2/stack/webstack"
	"github.com/pubgo/xerror"
)

var app = func() *chi.Mux {
	var route = chi.NewRouter()
	route.Use(middleware.Logger)
	route.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			if request.URL.Query().Get("stack") != "" {
				defer xerror.Resp(func(err xerror.XErr) {
					webstack.SnapshotHandler(writer, request)
				})
			} else {
				defer xerror.RespHttp(writer, request)
			}

			handler.ServeHTTP(writer, request)
		})
	})

	return route
}()

func Mux() *chi.Mux {
	return app
}
