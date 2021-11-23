package nsqc

import (
	"time"
)

var Name = "nsq"
var cfgList = make(map[string]Cfg)

type Cfg struct {
	Name           string        `json:"name"`
	Topic          string        `json:"topic"`
	Channel        string        `json:"channel"`
	Address        string        `json:"address"`
	Lookup         []string      `json:"lookup"`
	MaxInFlight    int           `json:"max_in_flight"`
	MaxConcurrency int           `json:"max_concurrency"`
	DialTimeout    time.Duration `json:"dial_timeout"`
	ReadTimeout    time.Duration `json:"read_timeout"`
	WriteTimeout   time.Duration `json:"write_timeout"`
	DrainTimeout   time.Duration `json:"drain_timeout"`
}

func (t Cfg) Build() (c *Client, err error) {
	return &Client{cfg: t}, nil
}

func GetDefaultCfg() Cfg {
	return Cfg{
		Address: "localhost:4150",
	}
}
