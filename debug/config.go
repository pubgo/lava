package debug

import (
	chiS "github.com/go-chi/chi/v5"
	"github.com/pubgo/lug/service/chi"
)

const Name = "debug"

var cfg = GetDefaultCfg()
var appMux *chiS.Mux

type Cfg struct {
	chi.Cfg
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Cfg: chi.GetDefaultCfg(),
	}
}
