package grpcc

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/xerror"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/core/middleware"

	_ "github.com/pubgo/lava/clients/grpcc/grpcc_lb/p2c"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
)

func init() {
	dix.Register(func(c config.Config, middlewares []middleware.Middleware) map[string]*Client {
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
