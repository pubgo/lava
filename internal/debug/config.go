package debug

import (
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"time"
)

type Cfg struct {
	Timeout   time.Duration `json:"timeout"`
	Logger    bool          `json:"logger"`
	Recover   bool          `json:"recover"`
	RequestID bool          `json:"req_id"`
	srv       *chi.Mux
}

func (t *Cfg) Get() *chi.Mux {
	if t.srv == nil {
		panic("please init chi")
	}

	return t.srv
}

func (t *Cfg) Build() error {
	var app = chi.NewMux()

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

	t.srv = app

	return nil
}

var cfg = Cfg{
	Timeout:   60 * time.Second,
	Logger:    true,
	Recover:   true,
	RequestID: true,
}
