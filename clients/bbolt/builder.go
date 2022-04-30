package bbolt

import (
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
)

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		cfg := cfgMap[name]
		inject.Register(fx.Provide(fx.Annotated{
			Name: inject.Name(name),
			Target: func(log *logging.Logger) *Client {
				return &Client{DB: cfg.Create(), log: log.Named(Name)}
			},
		}))
	}
}
