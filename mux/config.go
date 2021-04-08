package mux

import (
	"github.com/pubgo/golug/service/chiSrv"
)

const Name = "mux"

var cfg = GetDefaultCfg()

type Cfg struct {
	chiSrv.Cfg
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Cfg: chiSrv.GetDefaultCfg(),
	}
}
