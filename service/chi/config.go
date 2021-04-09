package chi

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"time"
)

type Cfg struct {
	Timeout   time.Duration `json:"timeout"`
	Logger    bool          `json:"logger"`
	Recover   bool          `json:"recover"`
	RequestID bool          `json:"request_id"`
}

func (t Cfg) Build() *chi.Mux {
	var app = chi.NewRouter()

	if t.Logger {
		app.Use(middleware.Logger)
	}

	if t.Recover {
		app.Use(middleware.Recoverer)
	}

	if t.RequestID {
		app.Use(middleware.RequestID)
	}

	if t.Timeout > 0 {
		app.Use(middleware.Timeout(t.Timeout))
	}

	return app
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Timeout:   60 * time.Second,
		Logger:    true,
		Recover:   true,
		RequestID: true,
	}
}
