package bbolt

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/lava/module"
)

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		if name == consts.KeyDefault {
			name = ""
		}

		cfg := cfgMap[name]
		module.Provide(fx.Annotated{
			Name: name,
			Target: func(log *logging.Logger) *Client {
				return &Client{DB: cfg.Create(), log: log.Named(Name)}
			},
		})
	}
}
