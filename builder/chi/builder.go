package chi

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Builder struct {
	srv *chi.Mux
}

func (t *Builder) Get() *chi.Mux {
	if t.srv == nil {
		panic("please init chi")
	}

	return t.srv
}

func (t *Builder) Build(cfg Cfg) error {
	var app = chi.NewRouter()

	if cfg.Logger {
		app.Use(middleware.Logger)
	}

	if cfg.Recover {
		app.Use(middleware.Recoverer)
	}

	if cfg.RequestID {
		app.Use(middleware.RequestID)
	}

	if cfg.Timeout > 0 {
		app.Use(middleware.Timeout(cfg.Timeout))
	}

	t.srv = app

	return nil
}

func New() Builder {
	return Builder{}
}
