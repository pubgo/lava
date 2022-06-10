package grpcc

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/config"

	_ "github.com/pubgo/lava/clients/grpcc/grpcc_lb/p2c"
)

func init() {
	dix.Register(func(c config.Config) map[string]*Client {
		var clients = make(map[string]*Client)
		var cfgMap = make(map[string]*grpcc_config.Cfg)
		xerror.Panic(c.Decode(grpcc_config.Name, cfgMap))
		for name := range cfgMap {
			clients[name] = NewClient(name, cfgMap[name])
		}
		return clients
	})
}
