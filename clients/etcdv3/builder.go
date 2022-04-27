package etcdv3

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/consts"
	"github.com/pubgo/lava/module"
)

const Name = "etcdv3"

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		if name == consts.KeyDefault {
			name = ""
		}

		cfg := cfgMap[name]
		module.Register(fx.Provide(fx.Annotated{
			Name: name,
			Target: func() *Client {
				return &Client{Client: cfg.Build()}
			},
		}))
	}
}
