package mux

import (
	"github.com/pubgo/lug/service/chi"
)

const Name = "mux"

var cfg = GetDefaultCfg()

type Cfg struct {
	chi.Cfg
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Cfg: chi.GetDefaultCfg(),
	}
}
