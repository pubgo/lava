package debug

import (
	cc "github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/builder/chi"
)

const Name = "debug"

var Addr = ":8081"

type Mux struct {
	*cc.Mux
	chi.Cfg
	srv chi.Builder
}
