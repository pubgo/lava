package grpcc

import (
	"github.com/pubgo/dix"
	"github.com/pubgo/funk/recovery"

	"github.com/pubgo/lava/clients/grpcc/grpcc_config"
	_ "github.com/pubgo/lava/clients/grpcc/grpcc_lb/p2c"
	"github.com/pubgo/lava/config"
	_ "github.com/pubgo/lava/core/registry/drivers/mdns"
	"github.com/pubgo/lava/service"
)

func init() {
	defer recovery.Exit()

	dix.Provider(func(c config.Config, middlewares []service.Middleware) map[string]*Client {
		return config.MakeClient(c, grpcc_config.Name, func(key string, cfg *grpcc_config.Cfg) *Client {
			cfg.Middlewares = middlewares
			return NewClient(key, cfg)
		})
	})
}
