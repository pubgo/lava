package bbolt

import (
	"github.com/pubgo/dix"

	"github.com/pubgo/lava/config"
	"github.com/pubgo/lava/logging"
)

func init() {
	dix.Register(func(c config.Config, log *logging.Logger) map[string]*Client {
		var cfgMap = config.Decode[*Cfg](c, Name)
		var clients = make(map[string]*Client)
		for name, cfg := range cfgMap {
			clients[name] = New(cfg.Create(), log.Named(Name))
		}
		return clients
	})
}
