package etcdv3

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/xerror"
)

const Name = "etcdv3"

func init() {
	dix.Register(func(c config.Config) map[string]*Client {
		var clients = make(map[string]*Client)
		var cfgMap = make(map[string]*Cfg)
		xerror.Panic(c.Decode(Name, cfgMap))

		for name := range cfgMap {
			cfg := cfgMap[name]
			clients[name] = &Client{Client: cfg.Build()}
		}
		return clients
	})

}
