package chi

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"time"
)

type Cfg struct {
	Timeout time.Duration `json:"timeout"`
}

func (t Cfg) Build() *chi.Mux {
	var app = chi.NewRouter()

	app.Use(middleware.Logger)
	app.Use(middleware.Recoverer)
	app.Use(middleware.RequestID)
	app.Use(middleware.Timeout(t.Timeout))

	return app
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Timeout: 60 * time.Second,
	}
}
