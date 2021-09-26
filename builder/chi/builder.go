package chi

import (
	"github.com/go-chi/chi/v5"

	"net/http"
	"sync"
)

type Builder struct {
	srv        *chi.Mux
	defaultMap sync.Map
}

func (t *Builder) Get() *chi.Mux {
	if t.srv == nil {
		panic("please init chi")
	}

	return t.srv
}

func (t *Builder) Build(cfg Cfg) error {
	var app = chi.NewRouter()
	app.NotFound(func(writer http.ResponseWriter, request *http.Request) {
		http.DefaultServeMux.ServeHTTP(writer, request)

		var _, ok = t.defaultMap.LoadOrStore(request.RequestURI, nil)
		if !ok {
			app.Handle(request.RequestURI, http.DefaultServeMux)
			return
		}
	})

	// 加载系统默认ServeMux
	app.Use(func(handler http.Handler) http.Handler {
		return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			var h, p = http.DefaultServeMux.Handler(request)
			if p == "" {
				handler.ServeHTTP(writer, request)
				return
			}

			var val, ok = t.defaultMap.Load(p)
			if ok {
				val.(http.Handler).ServeHTTP(writer, request)
				return
			}

			h.ServeHTTP(writer, request)
		})
	})

	//if cfg.Logger {
	//	app.Use(middleware.Logger)
	//}
	//
	//if cfg.Recover {
	//	app.Use(middleware.Recoverer)
	//}
	//
	//if cfg.RequestID {
	//	app.Use(middleware.RequestID)
	//}
	//
	//if cfg.Timeout > 0 {
	//	app.Use(middleware.Timeout(cfg.Timeout))
	//}

	t.srv = app

	return nil
}

func New() Builder {
	return Builder{}
}
