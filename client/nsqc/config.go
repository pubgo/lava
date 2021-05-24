package nsqc

import (
	"github.com/pubgo/x/jsonx"
)

var Name = "nsq"
var cfgList = make(map[string]Cfg)

type Cfg struct {
	Name           string         `json:"name"`
	Topic          string         `json:"topic"`
	Channel        string         `json:"channel"`
	Address        string         `json:"address"`
	Lookup         []string       `json:"lookup"`
	MaxInFlight    int            `json:"max_in_flight"`
	MaxConcurrency int            `json:"max_concurrency"`
	DialTimeout    jsonx.Duration `json:"dial_timeout"`
	ReadTimeout    jsonx.Duration `json:"read_timeout"`
	WriteTimeout   jsonx.Duration `json:"write_timeout"`
	DrainTimeout   jsonx.Duration `json:"drain_timeout"`
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Address: "localhost:4150",
	}
}
