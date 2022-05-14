package etcdv3

import (
	"github.com/pubgo/xerror"
	"go.uber.org/fx"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/inject"
)

const Name = "etcdv3"

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		cfg := cfgMap[name]
		inject.Register(fx.Provide(fx.Annotated{
			Name: inject.Name(name),
			Target: func() *Client {
				return &Client{Client: cfg.Get()}
			},
		}))
		inject.Register(fx.Provide(fx.Annotated{
			Group: Name,
			Target: func() *Client {
				return &Client{Client: cfg.Get()}
			},
		}))
	}
}
