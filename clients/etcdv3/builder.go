package etcdv3

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/xerror"
)

const Name = "etcdv3"

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		cfg := cfgMap[name]
		inject.NameGroup(Name, name, func() *Client {
			return &Client{Client: cfg.Get()}
		})
	}
}
