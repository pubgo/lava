package bbolt

import (
	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/inject"
	"github.com/pubgo/lava/logging"
	"github.com/pubgo/xerror"
)

func init() {
	var cfgMap = make(map[string]*Cfg)
	xerror.Panic(config.Decode(Name, cfgMap))

	for name := range cfgMap {
		cfg := cfgMap[name]
		inject.NameGroup(Name, name, func(log *logging.Logger) *Client {
			return New(cfg.Get(), log.Named(Name))
		})
	}
}
