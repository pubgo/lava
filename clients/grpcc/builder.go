package grpcc

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	_ "github.com/pubgo/lava/clients/grpcc/grpcc_lb/p2c"
	"github.com/pubgo/lava/config"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	"github.com/pubgo/lava/service"
)

func init() {
	dix.Provider(func(c config.Config, middlewares []service.Middleware) map[string]*Client {
		var clients = make(map[string]*Client)
		var cfgMap = make(map[string]*grpcc_config.Cfg)
		xerror.Panic(c.Decode(grpcc_config.Name, cfgMap))
		for name, cfg := range cfgMap {
			cfg.Middlewares = middlewares
			clients[name] = NewClient(name, cfg)
		}
		return clients
	})
}
