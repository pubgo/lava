package healthy

import (
	"github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/debug"

	"net/http"
)

var healthList []func() error

func init() {
	debug.On(func(mux *chi.Mux) {
		mux.Get("/health", func(writer http.ResponseWriter, request *http.Request) {
			for i := range healthList {
				if err := healthList[i](); err != nil {
					writer.WriteHeader(http.StatusBadRequest)
					writer.Write([]byte(err.Error()))
				}
			}

			writer.Write([]byte("ok"))
		})
	})
}
