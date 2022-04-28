package ossc

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/module"
)

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		cfg := cfgMap[name]
		module.Register(fx.Provide(fx.Annotated{
			Name: module.Name(name),
			Target: func(log *logging.Logger) *Client {
				return &Client{Client: cfg.Build()}
			},
		}))
	}
}
