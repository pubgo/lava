package debug

import (
	cc "github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/builder/chi"
	"github.com/pubgo/xlog"
)

const Name = "debug"
var logs=xlog.GetLogger(Name)

var Addr = ":8081"

type Mux struct {
	*cc.Mux
	chi.Cfg
	srv chi.Builder
}
