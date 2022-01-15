package mux

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/maruel/panicparse/v2/stack/webstack"
	"github.com/pubgo/xerror"
	"github.com/rs/cors"
)

var app = func() *chi.Mux {
	var route = chi.NewRouter()
	route.Use(middleware.Logger)
	route.Use(middleware.Recoverer)
	route.Use(middleware.RequestID)

	// 跨域处理
	route.Use(cors.Default().Handler)
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

func App() *chi.Mux {
	return app
}

func Use(middlewares ...func(http.Handler) http.Handler)             { app.Use(middlewares...) }
func With(middlewares ...func(http.Handler) http.Handler) chi.Router { return app.With(middlewares...) }
func Group(fn func(r chi.Router)) chi.Router                         { return app.Group(fn) }
func Route(pattern string, fn func(r chi.Router)) chi.Router         { return app.Route(pattern, fn) }
func Mount(pattern string, h http.Handler)                           { app.Mount(pattern, h) }
func Handle(pattern string, h http.Handler)                          { app.Handle(pattern, h) }
func HandleFunc(pattern string, h http.HandlerFunc)                  { app.HandleFunc(pattern, h) }
func Method(method, pattern string, h http.Handler)                  { app.Method(method, pattern, h) }
func MethodFunc(method, pattern string, h http.HandlerFunc)          { app.MethodFunc(method, pattern, h) }
func Connect(pattern string, h http.HandlerFunc)                     { app.Connect(pattern, h) }
func Delete(pattern string, h http.HandlerFunc)                      { app.Delete(pattern, h) }
func Get(pattern string, h http.HandlerFunc)                         { app.Get(pattern, h) }
func DebugGet(pattern string, h http.HandlerFunc)                    { app.Get(DebugPrefix(pattern), h) }
func Head(pattern string, h http.HandlerFunc)                        { app.Head(pattern, h) }
func Options(pattern string, h http.HandlerFunc)                     { app.Options(pattern, h) }
func Patch(pattern string, h http.HandlerFunc)                       { app.Patch(pattern, h) }
func Post(pattern string, h http.HandlerFunc)                        { app.Post(pattern, h) }
func Put(pattern string, h http.HandlerFunc)                         { app.Put(pattern, h) }
func Trace(pattern string, h http.HandlerFunc)                       { app.Trace(pattern, h) }
func NotFound(h http.HandlerFunc)                                    { app.NotFound(h) }
func MethodNotAllowed(h http.HandlerFunc)                            { app.MethodNotAllowed(h) }
