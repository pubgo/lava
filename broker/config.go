package broker

import (
	"github.com/pubgo/golug/consts"
)

var Name = "broker"

type Cfg struct {
	Driver string `json:"driver"`
	Name   string `json:"name"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Driver: "nsq",
		Name:   consts.Default,
	}
}
